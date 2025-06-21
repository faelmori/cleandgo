package types

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"

	it "github.com/rafa-mori/cleandgo/interfaces"
	gl "github.com/rafa-mori/cleandgo/logger"
	utl "github.com/rafa-mori/cleandgo/utils"
	l "github.com/rafa-mori/logz"
)

type FileTree struct {
	*Mutexes
	Logger             l.Logger             `json:"logger" yaml:"logger" xml:"logger" toml:"logger" gorm:"omitempty,logger"`                // Logger para registrar eventos
	PrintTree          bool                 `json:"printTree" yaml:"printTree" xml:"printTree" toml:"printTree" gorm:"omitempty,printTree"` // Indica se a árvore deve ser impressa
	TreeFileSource     string               `json:"treeFileSource" yaml:"treeFileSource" xml:"treeFileSource" toml:"treeFileSource" gorm:"omitempty,treeFileSource"`
	ComposerTargetPath string               `json:"composerTargetPath" yaml:"composerTargetPath" xml:"composerTargetPath" toml:"composerTargetPath" gorm:"omitempty,composerTargetPath"`
	EntriesMapOrigin   map[string]uuid.UUID `json:"entriesMapOrigin" yaml:"entriesMapOrigin" xml:"entriesMapOrigin" toml:"entriesMapOrigin" gorm:"omitempty,entriesMapOrigin"` // Mapa de origem das entradas
	Entries            []it.IFileEntry      `json:"entries" yaml:"entries" xml:"entries" toml:"entries" gorm:"omitempty,entries"`                                              // Lista de entradas de arquivo
	RootID             uuid.UUID            `json:"rootId" yaml:"rootId" xml:"rootId" toml:"rootId" gorm:"type:uuid,default:uuid_generate_v4()"`                               // ID do diretório raiz
	MaxDepth           int                  `json:"maxDepth" yaml:"maxDepth" xml:"maxDepth" toml:"maxDepth" gorm:"omitempty,maxDepth"`                                         // Profundidade máxima da árvore
	DrawedMap          map[string]string    `json:"drawed" yaml:"drawed" xml:"drawed" toml:"drawed" gorm:"omitempty,drawed"`                                                   // Mapa de símbolos usados para desenhar a árvore
	DirectoriesIcons   []string             `json:"directoriesIcons" yaml:"directoriesIcons" xml:"directoriesIcons" toml:"directoriesIcons" gorm:"omitempty,directoriesIcons"` // Ícones para diretórios
	FilesIcons         []string             `json:"filesIcons" yaml:"filesIcons" xml:"filesIcons" toml:"filesIcons" gorm:"omitempty,filesIcons"`                               // Ícones para arquivos
}

func NewFileTree(treeFileSource, composerTargetPath string, printTree bool, logger l.Logger, debug bool) (it.IFileTree, error) {
	// Logger resilient initialization
	if logger == nil {
		logger = l.GetLogger("CleandGO")
	}
	// Set debug mode for the logger
	gl.SetDebug(true)
	// Log
	gl.Log("debug", fmt.Sprintf("Initializing FileTree for path: %s", composerTargetPath))

	if composerTargetPath == "" && !printTree {
		gl.Log("error", "Path cannot be empty")
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Tree file source MUST exist and be a valid path (if will not be printed)
	if treeFileSource == "" && !printTree {
		gl.Log("error", "Tree file cannot be empty")
		return nil, fmt.Errorf("tree file cannot be empty")
	}
	if !filepath.IsAbs(treeFileSource) {
		if absPath, err := filepath.Abs(treeFileSource); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to get absolute path: %s", err))
			return nil, fmt.Errorf("failed to get absolute path: %s", err)
		} else {
			treeFileSource = absPath
		}
	}
	if _, statErr := os.Stat(treeFileSource); statErr != nil {
		gl.Log("error", fmt.Sprintf("Tree file does not exist: %s", treeFileSource))
		return nil, fmt.Errorf("tree file does not exist: %s", treeFileSource)
	}

	// Composer target path does't need to exist, because we will create it in the composer flow.
	if !filepath.IsAbs(composerTargetPath) {
		if absPath, err := filepath.Abs(composerTargetPath); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to get absolute path: %s", err.Error()))
			return nil, fmt.Errorf("failed to get absolute path: %s", err.Error())
		} else {
			composerTargetPath = absPath
		}
	}

	fte := &FileTree{
		Mutexes:            NewMutexesType(),
		PrintTree:          printTree,
		TreeFileSource:     treeFileSource,
		ComposerTargetPath: composerTargetPath,
		EntriesMapOrigin:   make(map[string]uuid.UUID), // Inicializa o mapa de origem das entradas
		Entries:            make([]it.IFileEntry, 0),
		RootID:             uuid.Nil, // Inicializa o ID do diretório raiz como vazio
		Logger:             logger,
		MaxDepth:           0, // Inicializa a profundidade máxima como 0
		DrawedMap: map[string]string{
			"├─ ": "H_LINE",
			"── ": "H_LINE_END",

			"├── ":  "V_LINE_END",
			"│   ":  "V_LINE_CONT",
			"└── ":  "V_LINE_LAST",
			"  ":    "V_LINE_SPACE_2",
			"   ":   "V_LINE_SPACE_3",
			"    ":  "V_LINE_SPACE_4",
			"     ": "V_LINE_SPACE_5",

			"├": "V_LINE_INIT",
			"│": "V_LINE_CONT_SINGLE",
			"└": "V_LINE_LAST_SINGLE",

			"	": "V_LINE_TAB",
		},
		DirectoriesIcons: []string{"📂", "📁", "🗂"},
		FilesIcons:       []string{"📜", "🔖", "🔥", "✔"},
	}

	if err := fte.ParseTree(); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to parse tree source: %s", err.Error()))
		return nil, fmt.Errorf("failed to parse tree source: %s", err.Error())
	}

	// Log the number of entries loaded
	gl.Log("debug", fmt.Sprintf("FileTree parsed with %d entries", len(fte.Entries)))

	return fte, nil
}

func (ft *FileTree) GetEntries() []it.IFileEntry {
	return ft.Entries
}
func (ft *FileTree) GetEntryByName(name string) it.IFileEntry {
	if name == "" {
		gl.Log("error", "Entry name cannot be empty")
		return nil
	}
	if entryID, exists := ft.EntriesMapOrigin[name]; exists {
		for _, entry := range ft.Entries {
			if entry.GetID() == entryID {
				return entry // Retorna a entrada correspondente ao nome
			}
		}
		gl.Log("warn", fmt.Sprintf("Entry with name '%s' not found in FileTree", name))
		return nil // Retorna nil se não encontrar a entrada
	}
	gl.Log("warn", fmt.Sprintf("Entry with name '%s' not found in EntriesMapOrigin", name))
	return nil // Retorna nil se o nome não existir no mapa
}
func (ft *FileTree) AddEntry(entry it.IFileEntry) {
	if entry == nil {
		gl.Log("error", "Cannot add nil entry to FileTree")
		return
	}

	// Mapeia o nome da entrada para o ID
	ft.EntriesMapOrigin[entry.GetName()] = entry.GetID()

	ft.Entries = append(ft.Entries, entry)

	if (ft.RootID == uuid.Nil || ft.RootID.String() == "00000000-0000-0000-0000-000000000000") && entry.GetDepth() == 0 {
		// Se o ID do diretório raiz ainda não estiver definido,
		// define o primeiro diretório recebido como raiz
		// Isso garante que o primeiro diretório adicionado seja sempre o raiz
		ft.RootID = entry.GetID()
	}

	go ft.SetEntriesDepth() // Define a profundidade das entradas em segundo plano
}
func (ft *FileTree) SetEntriesDepth() {
	var maxDepth int
	if ft.Mutexes == nil {
		ft.Mutexes = NewMutexesType() // Inicializa Mutexes se for nil
	}

	ft.MuLock()
	defer ft.MuUnlock()

	depthRegex := regexp.MustCompile(`^([\s│├└──]*)`)
	for i, entry := range ft.Entries {
		matches := depthRegex.FindStringSubmatch(entry.GetOriginName())
		if len(matches) > 0 {
			depth := strings.Count(matches[1], "│") + strings.Count(matches[1], "├") + strings.Count(matches[1], "└")
			ft.Entries[i].SetDepth(depth)
			if depth > maxDepth {
				maxDepth = depth
			}
		} else {
			ft.Entries[i].SetDepth(0)
		}
	}
}
func (ft *FileTree) GetEntryByID(id uuid.UUID) it.IFileEntry {
	for _, entry := range ft.Entries {
		if entry.GetID() == id {
			return entry
		}
	}
	return nil // Retorna nil se não encontrar
}
func (ft *FileTree) GetChildren(parentID uuid.UUID) []it.IFileEntry {
	var children []it.IFileEntry
	for _, entry := range ft.Entries {
		if entry.GetParentID() == parentID {
			children = append(children, entry)
		}
	}
	return children // Retorna a lista de filhos
}
func (ft *FileTree) Sanitize(dirtyData []byte) error {
	// 1: Remove entradas inválidas ou duplicadas
	validEntries := make([]it.IFileEntry, 0)
	seenIDs := make(map[uuid.UUID]bool)

	for _, entry := range ft.Entries {
		if entry.GetID() == uuid.Nil || seenIDs[entry.GetID()] {
			continue // Ignora entradas sem ID ou duplicadas
		}
		seenIDs[entry.GetID()] = true
		validEntries = append(validEntries, entry)
	}

	ft.Entries = validEntries // Atualiza a lista de entradas
	gl.Log("debug", fmt.Sprintf("Sanitized FileTree with %d valid entries", len(ft.Entries)))

	// 2: Remove todo conteúdo extra e residual do arquivo de árvore
	if len(dirtyData) > 0 {
		// Cria um novo arquivo temporário para armazenar os dados limpos
		tempFile, err := os.CreateTemp("", "cleandgo_tree_*.json")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to create temporary file: %s", err.Error()))
			return fmt.Errorf("failed to create temporary file: %s", err.Error())
		}
		defer os.Remove(tempFile.Name()) // Remove o arquivo temporário após o uso

		// Escreve os dados limpos no arquivo temporário
		if _, err := tempFile.Write(dirtyData); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to write to temporary file: %s", err.Error()))
			return fmt.Errorf("failed to write to temporary file: %s", err.Error())
		}

		gl.Log("debug", fmt.Sprintf("Sanitized data written to temporary file: %s", tempFile.Name()))

		// Faz backup do arquivo original
		backupFile := fmt.Sprintf("%s.bak", ft.TreeFileSource)
		if err := os.Rename(ft.TreeFileSource, backupFile); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to backup original file: %s", err.Error()))
			return fmt.Errorf("failed to backup original file: %s", err.Error())
		}
		gl.Log("debug", fmt.Sprintf("Backup of original file created: %s", backupFile))

		// Move o arquivo temporário para o local original
		if err := os.Rename(tempFile.Name(), ft.TreeFileSource); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to move temporary file to original location: %s", err.Error()))
			return fmt.Errorf("failed to move temporary file to original location: %s", err.Error())
		}
		gl.Log("debug", fmt.Sprintf("Temporary file moved to original location: %s", ft.TreeFileSource))

		// Recarrega os dados do arquivo de árvore sanitarizado
		mapper := NewMapper(&ft, ft.TreeFileSource)
		if obj, err := mapper.DeserializeFromFile("json"); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to deserialize sanitized file: %s", err))

			return fmt.Errorf("failed to deserialize sanitized file: %s", err.Error())
		} else {
			if obj != nil {
				ft = *obj
			}
		}
	}

	// 3: Valida os possíveis e prováveis tipos das entradas que ainda "não existem" fornecidos no arquivo de árvore
	for i, entry := range ft.Entries {
		if entry.GetType() != "file" && entry.GetType() != "directory" {
			extension := filepath.Ext(entry.GetName())
			if extension == "" {
				slashes := strings.HasSuffix(entry.GetName(), "/") || strings.HasSuffix(entry.GetName(), "\\")
				if slashes {
					ft.Entries[i].SetType("directory") // Se termina com barra, é um diretório
				} else {
					ft.Entries[i].SetType("file") // Caso contrário, é um arquivo
				}
			} else {
				ft.Entries[i].SetType("file") // Se tem extensão, é um arquivo
			}
		}
	}

	// 4: Verifica se o ID do diretório raiz está definido
	if ft.RootID == uuid.Nil || ft.RootID.String() == "00000000-0000-0000-0000-000000000000" {
		// Se não estiver definido, define o primeiro diretório como raiz
		for _, entry := range ft.Entries {
			if entry.GetType() == "directory" {
				ft.RootID = entry.GetID()
				gl.Log("debug", fmt.Sprintf("Root ID set to: %s", ft.RootID))
				break
			}
		}
		if ft.RootID == uuid.Nil {
			gl.Log("error", "No root directory found in FileTree entries")
			return fmt.Errorf("no root directory found in FileTree entries")
		}
	}
	gl.Log("debug", "FileTree sanitized successfully")

	return nil
}
func (ft *FileTree) SetMaxDepth(depth int) {
	ft.MaxDepth = depth
}
func (ft *FileTree) GetRootID() uuid.UUID {
	return ft.RootID
}
func (ft *FileTree) SetRootID(id uuid.UUID) {
	ft.RootID = id
}

func (ft *FileTree) ParseTree() error {
	// Get tree source and composer target path
	treeFileSource := ft.TreeFileSource
	if treeFileSource == "" && !ft.PrintTree {
		gl.Log("fatal", "Tree file source cannot be empty")
	}
	// Get composer target path
	composerTargetPath := ft.ComposerTargetPath
	if composerTargetPath == "" && !ft.PrintTree {
		gl.Log("fatal", "Composer target path cannot be empty")
	}

	// First parse is directly from the tree file source
	// The others will be through the IMapper interface
	// Read the tree file and populate the entries
	if treeFileSource != "" && !ft.PrintTree {
		gl.Log("debug", fmt.Sprintf("Loading tree file from source: %s", treeFileSource))

		file, err := os.Open(treeFileSource)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to open tree file: %s", err))
			return nil
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)

		// Scanner para ler o arquivo linha por linha
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue // Ignora linhas vazias
			}
			// Parse the line into a FileEntry
			if entry, err := ParseFieldsFromTreeView(line, ft); err != nil {
				gl.Log("error", fmt.Sprintf("Failed to parse line '%s': %s", line, err))
				return fmt.Errorf("failed to parse line '%s': %s", line, err)
			} else {
				ft.AddEntry(entry) // Adiciona a entrada ao FileTree
			}
		}

		// Check for errors during scanning
		if err := scanner.Err(); err != nil {
			gl.Log("error", fmt.Sprintf("Error reading tree file: %s", err))
			return nil
		}

		gl.Log("debug", fmt.Sprintf("Loaded %d entries from tree file: %s", len(ft.Entries), treeFileSource))

		// Set the deepness of the entries based on their structure
		if err := utl.SetTreeViewEntriesDeepness(ft); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to set tree view entries deepness: %s", err))
			return fmt.Errorf("failed to set tree view entries deepness: %s", err)
		}
	} else {
		gl.Log("debug", "No tree file provided, initializing empty FileTree")
	}
	return nil
}
func (ft *FileTree) SerializeToFile(format string) error {
	// Do a backup before serializing the new content.
	if err := ft.BackupTreeFile(); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to backup tree file: %s", err))
		return fmt.Errorf("failed to backup tree file: %s", err)
	}

	// Serializes the FileTree to a file in the specified format
	mapper := NewMapper(ft, ft.TreeFileSource)
	mapper.SerializeToFile(format)

	// Check if the file was created successfully
	if !utl.CheckFileExists(ft.TreeFileSource) {
		// If the file does not exist after serialization, log an error and restore the backup
		gl.Log("error", fmt.Sprintf("Failed to serialize FileTree to file: %s", ft.TreeFileSource))
		if err := ft.RestoreTreeFile(); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to restore backup file: %s", err))
			return fmt.Errorf("failed to restore backup file: %s", err)
		}
		return fmt.Errorf("failed to serialize FileTree to file: %s", ft.TreeFileSource)
	}

	return nil
}
func (ft *FileTree) LoadFromFile(format string) error {
	// Load the FileTree from a file in the specified format
	mapper := NewMapper(ft, ft.TreeFileSource)
	if obj, err := mapper.DeserializeFromFile(format); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to deserialize FileTree from file: %s", err))
		return fmt.Errorf("failed to deserialize FileTree from file: %s", err)
	} else {
		if obj != nil {
			ft = obj // Update the FileTree with the loaded data
		}
	}

	// Set the deepness of the entries based on their structure
	if err := utl.SetTreeViewEntriesDeepness(ft); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to set tree view entries deepness: %s", err))
		return fmt.Errorf("failed to set tree view entries deepness: %s", err)
	}

	return nil
}
func (ft *FileTree) BackupTreeFile() error {
	// Do a backup before serializing the new content.
	if utl.CheckFileExists(ft.TreeFileSource) {
		backupFile := fmt.Sprintf("%s.bak", ft.TreeFileSource)
		if err := os.Rename(ft.TreeFileSource, backupFile); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to create backup of tree file: %s", err))
			return fmt.Errorf("failed to create backup of tree file: %s", err)
		}
		gl.Log("debug", fmt.Sprintf("Backup of tree file created: %s", backupFile))
	} else {
		gl.Log("debug", fmt.Sprintf("No existing tree file to backup: %s", ft.TreeFileSource))
	}
	return nil
}
func (ft *FileTree) RestoreTreeFile() error {
	if utl.CheckFileExists(fmt.Sprintf("%s.bak", ft.TreeFileSource)) {
		gl.Log("debug", fmt.Sprintf("Restoring backup file: %s.bak", ft.TreeFileSource))
		backupFile := fmt.Sprintf("%s.bak", ft.TreeFileSource)
		if restoreErr := os.Rename(backupFile, ft.TreeFileSource); restoreErr != nil {
			gl.Log("error", fmt.Sprintf("Failed to restore backup file: %s", restoreErr))
			return fmt.Errorf("failed to restore backup file: %s", restoreErr)
		}
		gl.Log("debug", fmt.Sprintf("Backup file restored: %s", ft.TreeFileSource))
	} else {
		gl.Log("error", fmt.Sprintf("No backup file found to restore: %s.bak", ft.TreeFileSource))
	}
	return nil
}
func (ft *FileTree) GetDirectoriesIcons() []string {
	return ft.DirectoriesIcons
}

func (ft *FileTree) GetFilesIcons() []string {
	return ft.FilesIcons
}
func (ft *FileTree) GetDrawedMap() map[string]string {
	return ft.DrawedMap
}
func (ft *FileTree) GetFileTreeType() any {
	return ft
}

func ParseFieldsFromTreeView(line string, ft it.IFileTree) (it.IFileEntry, error) {
	// Verifica se é um diretório ou arquivo pelo último campo
	entryType := "unknown"
	// if fields[len(fields)-1] == "file" || fields[len(fields)-1] == "directory" {
	// 	entryType = fields[len(fields)-1]
	// }

	lineEntry, comments := utl.ExtractComment(strings.TrimSpace(strings.ToValidUTF8(line, ""))) // Extrai o comentário, se houver

	// Verifica se a linha contém os ícones de identificação de diretórios e arquivos,
	// se sim, já determina o tipo de entrada e remove os ícones
	if utl.ContainsIcon(lineEntry, ft.GetDirectoriesIcons()) {
		entryType = "directory" // Se contém ícone de diretório, é um diretório
	} else if utl.ContainsIcon(lineEntry, ft.GetFilesIcons()) {
		entryType = "file" // Se contém ícone de arquivo, é um arquivo
	} else if strings.HasSuffix(lineEntry, "/") || strings.HasSuffix(lineEntry, "\\") {
		// Se termina com barra, é um diretório, SEMPRE
		entryType = "directory"
	} else if strings.Contains(lineEntry, ".") && !strings.HasSuffix(lineEntry, "/") && !strings.HasSuffix(lineEntry, "\\") && filepath.Ext(lineEntry) != "" {
		// Se contém ponto e não termina com barra, é um arquivo
		// Não vamos considerar arquivos ocultos sem extensão, eles serão tratados como diretórios
		entryType = "file"
	} else if filepath.Ext(lineEntry) != "" && !strings.HasSuffix(lineEntry, "/") && !strings.HasSuffix(lineEntry, "\\") {
		entryType = "unknown" // Caso contrário, é um tipo desconhecido
	}

	lineEntry = utl.SanitizeLineIcons(lineEntry, ft.GetDirectoriesIcons(), ft.GetFilesIcons()) // Remove ícones de arquivo

	// A ideia é buscar SEMPRE o último "bloco" válido de texto na linha, que venha depois de / ou \\,
	// e que corresponda corretamente ao tipo de entrada. Diretórios não precisam necessariamente de ter a árvore com os paths
	// no nome, eles podem ser "construídos" posteriormente com o deepness e a estrutura de árvore.
	// Isso porque nem sempre eles irão existir previamente, e nem sempre eles terão um nome totalmente válido sem espaços e outras coisas bizarras.
	name := strings.TrimSuffix(strings.TrimSuffix(lineEntry, "/"), "\\")

	// Remove ícones e caracteres extras do nome
	name = utl.SanitizeLineIcons(name, ft.GetDirectoriesIcons(), ft.GetFilesIcons()) // Remove ícones e caracteres extras

	// Remove caracteres inválidos do nome
	name = utl.SanitizeLineChars(name) // Remove a barra final, se existir

	// Remove os símbolos de estrutura da árvore do nome
	name = utl.RemoveDrawedIdentifiers(name, ft.GetDrawedMap()) // Remove os símbolos de estrutura da árvore do nome

	// Antes de criar a entrada, precisamos garantir que o nome esteja correto
	name = strings.TrimSpace(name) // Remove espaços extras ao redor

	// Nome original da linha, sem espaços extras
	originName := strings.TrimSpace(strings.ToValidUTF8(line, ""))

	// Se o nome estiver vazio, não cria a entrada
	if name == "" {
		gl.Log("warn", fmt.Sprintf("Invalid entry name from line '%s'", line))
		return nil, nil
	}

	// Remove a estrutura da árvore de diretórios do nome, se existir
	// Pega o último bloco após a última barra
	name = strings.Split(name, "/")[len(strings.Split(name, "/"))-1]

	// Pega o último bloco após a última barra invertida
	name = strings.Split(name, "\\")[len(strings.Split(name, "\\"))-1]

	// Cria a entrada de arquivo
	if entry, entryErr := NewFileEntry(
		uuid.New(), // Gera um novo UUID para a entrada
		uuid.Nil,   // Inicialmente sem pai, será definido posteriormente
		entryType,
		name,
		originName,
		0,
		0,
		comments,
	); entryErr != nil {
		gl.Log("error", fmt.Sprintf("Failed to create FileEntry from line '%s': %s", line, entryErr))
		return nil, fmt.Errorf("failed to create FileEntry from line '%s': %s", line, entryErr)
	} else {
		return entry, nil
	}
}
