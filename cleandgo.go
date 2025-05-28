package cleandgo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gl "github.com/faelmori/cleandgo/logger"
	t "github.com/faelmori/cleandgo/types"
	l "github.com/faelmori/logz"
	"github.com/google/uuid"
)

type FileTree struct {
	*t.Mutexes
	Logger             l.Logger             `json:"logger" yaml:"logger" xml:"logger" toml:"logger" gorm:"omitempty,logger"`                // Logger para registrar eventos
	PrintTree          bool                 `json:"printTree" yaml:"printTree" xml:"printTree" toml:"printTree" gorm:"omitempty,printTree"` // Indica se a árvore deve ser impressa
	TreeFileSource     string               `json:"treeFileSource" yaml:"treeFileSource" xml:"treeFileSource" toml:"treeFileSource" gorm:"omitempty,treeFileSource"`
	ComposerTargetPath string               `json:"composerTargetPath" yaml:"composerTargetPath" xml:"composerTargetPath" toml:"composerTargetPath" gorm:"omitempty,composerTargetPath"`
	EntriesMapOrigin   map[string]uuid.UUID `json:"entriesMapOrigin" yaml:"entriesMapOrigin" xml:"entriesMapOrigin" toml:"entriesMapOrigin" gorm:"omitempty,entriesMapOrigin"` // Mapa de origem das entradas
	Entries            []*t.FileEntry       `json:"entries" yaml:"entries" xml:"entries" toml:"entries" gorm:"omitempty,entries"`                                              // Lista de entradas de arquivo
	RootID             uuid.UUID            `json:"rootId" yaml:"rootId" xml:"rootId" toml:"rootId" gorm:"type:uuid,default:uuid_generate_v4()"`                               // ID do diretório raiz
	MaxDepth           int                  `json:"maxDepth" yaml:"maxDepth" xml:"maxDepth" toml:"maxDepth" gorm:"omitempty,maxDepth"`                                         // Profundidade máxima da árvore
	DrawedMap          map[string]string    `json:"drawed" yaml:"drawed" xml:"drawed" toml:"drawed" gorm:"omitempty,drawed"`                                                   // Mapa de símbolos usados para desenhar a árvore
}

func NewFileTree(treeFileSource, composerTargetPath string, printTree bool, logger l.Logger, debug bool) (*FileTree, error) {
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
			gl.Log("error", fmt.Sprintf("Failed to get absolute path: %s", err.Error()))
			return nil, fmt.Errorf("failed to get absolute path: %s", err.Error())
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
		Mutexes:            t.NewMutexesType(),
		PrintTree:          printTree,
		TreeFileSource:     treeFileSource,
		ComposerTargetPath: composerTargetPath,
		EntriesMapOrigin:   make(map[string]uuid.UUID), // Inicializa o mapa de origem das entradas
		Entries:            make([]*t.FileEntry, 0),
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
	}

	if err := fte.ParseTree(); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to parse tree source: %s", err.Error()))
		return nil, fmt.Errorf("failed to parse tree source: %s", err.Error())
	}

	// Log the number of entries loaded
	gl.Log("debug", fmt.Sprintf("FileTree parsed with %d entries", len(fte.Entries)))

	return fte, nil
}

func (ft *FileTree) AddEntry(entry *t.FileEntry) {
	if entry == nil {
		gl.Log("error", "Cannot add nil entry to FileTree")
		return
	}

	// Mapeia o nome da entrada para o ID
	ft.EntriesMapOrigin[entry.Name] = entry.ID

	ft.Entries = append(ft.Entries, entry)

	if ft.RootID == uuid.Nil && entry.Type == "directory" {
		ft.RootID = entry.ID // Define o primeiro diretório como raiz
	}
}
func (ft *FileTree) GetEntryByID(id uuid.UUID) *t.FileEntry {
	for _, entry := range ft.Entries {
		if entry.ID == id {
			return entry
		}
	}
	return nil // Retorna nil se não encontrar
}
func (ft *FileTree) GetChildren(parentID uuid.UUID) []*t.FileEntry {
	var children []*t.FileEntry
	for _, entry := range ft.Entries {
		if entry.ParentID == parentID {
			children = append(children, entry)
		}
	}
	return children // Retorna a lista de filhos
}
func (ft *FileTree) Sanitize(dirtyData []byte) error {
	// 1: Remove entradas inválidas ou duplicadas
	validEntries := make([]*t.FileEntry, 0)
	seenIDs := make(map[uuid.UUID]bool)

	for _, entry := range ft.Entries {
		if entry.ID == uuid.Nil || seenIDs[entry.ID] {
			continue // Ignora entradas sem ID ou duplicadas
		}
		seenIDs[entry.ID] = true
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
		mapper := t.NewMapper(&ft, ft.TreeFileSource)
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
		if entry.Type != "file" && entry.Type != "directory" {
			extension := filepath.Ext(entry.Name)
			if extension == "" {
				slashes := strings.HasSuffix(entry.Name, "/") || strings.HasSuffix(entry.Name, "\\")
				if slashes {
					ft.Entries[i].Type = "directory" // Se termina com barra, é um diretório
				} else {
					ft.Entries[i].Type = "file" // Caso contrário, é um arquivo
				}
			} else {
				ft.Entries[i].Type = "file" // Se tem extensão, é um arquivo
			}
		}
	}

	// 4: Verifica se o ID do diretório raiz está definido
	if ft.RootID == uuid.Nil || ft.RootID.String() == "00000000-0000-0000-0000-000000000000" {
		// Se não estiver definido, define o primeiro diretório como raiz
		for _, entry := range ft.Entries {
			if entry.Type == "directory" {
				ft.RootID = entry.ID
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
			if entry, err := parseFieldsFromTreeView(line, ft); err != nil {
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
		if err := setTreeViewEntriesDeepness(ft); err != nil {
			gl.Log("error", fmt.Sprintf("Failed to set tree view entries deepness: %s", err))
			return fmt.Errorf("failed to set tree view entries deepness: %s", err)
		}
	} else {
		gl.Log("debug", "No tree file provided, initializing empty FileTree")
	}
	return nil
}

func parseFieldsFromTreeView(line string, ft *FileTree) (*t.FileEntry, error) {
	// Verifica se é um diretório ou arquivo pelo último campo
	entryType := "unknown"
	// if fields[len(fields)-1] == "file" || fields[len(fields)-1] == "directory" {
	// 	entryType = fields[len(fields)-1]
	// }

	// Nem todas as linhas terão /, nem todas as linhas terão o tipo de entrada explícito
	// Portanto devemos inserir raw, com tipo unknown e deepness 0
	// Permitindo que a entrada seja tratada posteriormente de forma mais adequada usando métodos auxiliares.
	if strings.HasSuffix(line, "/") || strings.HasSuffix(line, "\\") {
		// Se termina com barra, é um diretório, SEMPRE
		entryType = "directory"
	} else if strings.Contains(line, ".") && !strings.HasSuffix(line, "/") && !strings.HasSuffix(line, "\\") && filepath.Ext(line) != "" {
		// Se contém ponto e não termina com barra, é um arquivo
		// Não vamos considerar arquivos ocultos sem extensão, eles serão tratados como diretórios
		entryType = "file"
	} else if filepath.Ext(line) != "" && !strings.HasSuffix(line, "/") && !strings.HasSuffix(line, "\\") {
		entryType = "unknown" // Caso contrário, é um tipo desconhecido
	}

	// Nome original da linha, sem espaços extras
	originName := strings.TrimSpace(strings.ToValidUTF8(line, ""))

	// A ideia é buscar SEMPRE o último "bloco" válido de texto na linha, que venha depois de / ou \\,
	// e que corresponda corretamente ao tipo de entrada. Diretórios não precisam necessariamente de ter a árvore com os paths
	// no nome, eles podem ser "construídos" posteriormente com o deepness e a estrutura de árvore.
	// Isso porque nem sempre eles irão existir previamente, e nem sempre eles terão um nome totalmente válido sem espaços e outras coisas bizarras.
	name := strings.TrimSuffix(strings.TrimSuffix(originName, "/"), "\\")

	// Remove ícones e caracteres extras do nome
	name = sanitizeLineIcons(name) // Remove ícones e caracteres extras

	// Remove caracteres inválidos do nome
	name = sanitizeLineChars(name) // Remove a barra final, se existir

	// Antes de criar a entrada, precisamos garantir que o nome esteja correto
	name = strings.TrimSpace(name) // Remove espaços extras ao redor

	drawedMap := ft.DrawedMap
	if drawedMap == nil {
		gl.Log("error", "DrawedMap is nil, cannot parse line")
		return nil, fmt.Errorf("drawedMap is nil, cannot parse line '%s'", line)
	}
	// Substitui os caracteres de estrutura da árvore pelos símbolos correspondentes
	for char, symbol := range drawedMap {
		// Remove os símbolos de estrutura da árvore do nome
		name = strings.ReplaceAll(name, char, symbol)
	}

	// Se o nome estiver vazio, não cria a entrada
	if name == "" {
		gl.Log("error", fmt.Sprintf("Invalid entry name from line '%s'", line))
		return nil, nil
	}

	// Remove a estrutura da árvore de diretórios do nome, se existir
	// Pega o último bloco após a última barra
	name = strings.Split(name, "/")[len(strings.Split(name, "/"))-1]

	// Pega o último bloco após a última barra invertida
	name = strings.Split(name, "\\")[len(strings.Split(name, "\\"))-1]

	// Cria a entrada de arquivo
	if entry, entryErr := t.NewFileEntry(
		uuid.New(), // Gera um novo UUID para a entrada
		uuid.Nil,   // Inicialmente sem pai, será definido posteriormente
		entryType,
		name,
		originName,
		0,
		0,
	); entryErr != nil {
		gl.Log("error", fmt.Sprintf("Failed to create FileEntry from line '%s': %s", line, entryErr))
		return nil, fmt.Errorf("failed to create FileEntry from line '%s': %s", line, entryErr)
	} else {
		return entry, nil
	}
}

func sanitizeLineIcons(line string) string {
	unwantedChars := []string{"📂", "📁", "🗂", "📜", "🔖", "🔥", "✔"} // Ícones e caracteres extras
	for _, char := range unwantedChars {
		line = strings.ReplaceAll(line, char, "")
	}
	return strings.TrimSpace(line) // Remove espaços extras ao redor
}
func sanitizeLineChars(line string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_\/.-]`)
	return re.ReplaceAllString(line, "")
}
func setTreeViewEntriesDeepness(ft *FileTree) error {
	// Define a profundidade para cada entrada com base na sua posição na árvore (por barras, indentação e simbologia visual)
	//treeStructureSymbols := []string{"├── ", "│   ", "└── ", "    "}

	// Esse regex não tá eficiente. Bora mudar pra algo simislar ao drawed map, mas que capture a profundidade de forma mais precisa.
	// Regex para capturar a profundidade da entrada
	drawedMap := ft.DrawedMap
	if drawedMap == nil {
		gl.Log("error", "DrawedMap is nil, cannot set tree view entries deepness")
		return fmt.Errorf("drawedMap is nil, cannot set tree view entries deepness")
	}

	var maxDepth int
	// depthRegexStr := ""
	// // Cria uma string regex para capturar a profundidade
	// for char, symbol := range drawedMap {
	// 	if symbol == "V_LINE_SPACE_2" || symbol == "V_LINE_SPACE_3" || symbol == "V_LINE_SPACE_4" || symbol == "V_LINE_SPACE_5" {
	// 		// Adiciona os símbolos de espaço como profundidade
	// 		depthRegexStr += fmt.Sprintf(`(%s)`, regexp.QuoteMeta(char))
	// 	} else if symbol == "H_LINE" || symbol == "H_LINE_END" || symbol == "V_LINE_INIT" || symbol == "V_LINE_CONT_SINGLE" || symbol == "V_LINE_LAST_SINGLE" {
	// 		// Ignora os símbolos de linha horizontal e vertical que não indicam profundidade
	// 		continue
	// 	} else {
	// 		// Adiciona os outros símbolos como profundidade
	// 		depthRegexStr += fmt.Sprintf(`(%s)`, regexp.QuoteMeta(char))
	// 	}
	// }

	depthRegex := regexp.MustCompile(`^([\s│├└──]*)`)
	// if err != nil {
	// 	gl.Log("error", fmt.Sprintf("Failed to compile depth regex: %s", err))
	// 	return fmt.Errorf("failed to compile depth regex: %s", err)
	// }
	// depthRegexStr = strings.TrimSuffix(depthRegexStr, "|") // Remove o último pipe, se existir
	// depthRegex, err := regexp.Compile(depthRegexStr)
	// if err != nil {
	// 	gl.Log("error", fmt.Sprintf("Failed to compile depth regex: %s", err))
	// 	return fmt.Errorf("failed to compile depth regex: %s", err)
	// }

	for i, entry := range ft.Entries {
		matches := depthRegex.FindStringSubmatch(entry.OriginName)
		if len(matches) > 0 {
			depth := strings.Count(matches[1], "│") + strings.Count(matches[1], "├") + strings.Count(matches[1], "└")
			ft.Entries[i].Depth = depth
			if depth > maxDepth {
				maxDepth = depth
			}
		} else {
			ft.Entries[i].Depth = 0 // Se não houver correspondência, define como 0
		}
	}

	// Define a profundidade máxima para o FileTree
	ft.MaxDepth = maxDepth
	gl.Log("debug", fmt.Sprintf("Set tree view entries deepness with max depth: %d", maxDepth))

	foundRoot := false

	for _, entry := range ft.Entries {
		if entry.Type == "directory" && entry.Depth == 0 {
			ft.RootID = entry.ID
			foundRoot = true

			gl.Log("debug", fmt.Sprintf("Root ID set to: %s", ft.RootID))
			break
		}
	}

	if !foundRoot {
		gl.Log("error", "No root directory found in FileTree entries")
		// return fmt.Errorf("no root directory found in FileTree entries")
	}

	// Define o ID do diretório raiz, se ainda não estiver definido
	if ft.RootID == uuid.Nil {
		for _, entry := range ft.Entries {
			if entry.Type == "directory" {
				ft.RootID = entry.ID
				gl.Log("debug", fmt.Sprintf("Root ID set to: %s", ft.RootID))
				break
			}
		}
		if ft.RootID == uuid.Nil {
			gl.Log("error", "No root directory found in FileTree entries")
		}
	}

	gl.Log("debug", fmt.Sprintf("FileTree entries deepness set with %d entries", len(ft.Entries)))

	// Define os IDs de cada entrada com base na posição na lista
	if err := setTreeViewDrawedIdentifiers(ft); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to set tree view identifiers: %s", err))
		return fmt.Errorf("failed to set tree view identifiers: %s", err)
	}

	// Define as referências de estrutura para cada entrada
	if err := setTreeStructureReferences(ft); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to set tree structure references: %s", err))
		return fmt.Errorf("failed to set tree structure references: %s", err)
	}

	gl.Log("debug", "FileTree structure references set successfully")

	return nil
}
func setTreeViewDrawedIdentifiers(ft *FileTree) error {
	// Define os IDs de cada entrada com base na posição na lista
	for i := range ft.Entries {
		ft.Entries[i].ID = uuid.New() // Gera um novo UUID para cada entrada
	}

	// Define o ID do diretório raiz, se ainda não estiver definido
	if ft.RootID == uuid.Nil {
		for _, entry := range ft.Entries {
			if entry.Type == "directory" {
				ft.RootID = entry.ID
				break
			}
		}
	}

	gl.Log("debug", fmt.Sprintf("Root ID set to: %s", ft.RootID))

	if ft.RootID == uuid.Nil {
		gl.Log("error", "No root directory found in FileTree entries")
		return fmt.Errorf("no root directory found in FileTree entries")
	}

	gl.Log("debug", fmt.Sprintf("Set tree view drawed identifiers with %d entries", len(ft.Entries)))

	return nil
}
func setTreeStructureReferences(ft *FileTree) error {
	// Cria um mapa para facilitar a busca de entradas por ID
	entryMap := make(map[uuid.UUID]*t.FileEntry)

	for i := range ft.Entries {
		entryMap[ft.Entries[i].ID] = ft.Entries[i]
	}

	// Define as referências de estrutura para cada entrada
	for i := range ft.Entries {
		entry := ft.Entries[i]
		if entry.ParentID != uuid.Nil {
			if parent, ok := entryMap[entry.ParentID]; ok {
				entry.Parent = parent
			}
		}
	}
	return nil
}
