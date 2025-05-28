package cleandgo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gl "github.com/faelmori/cleandgo/logger"
	t "github.com/faelmori/cleandgo/types"
	l "github.com/faelmori/logz"
	"github.com/google/uuid"
)

type FileTree struct {
	*t.Mutexes
	Logger             l.Logger      `json:"logger" yaml:"logger" xml:"logger" toml:"logger" gorm:"omitempty,logger"`                // Logger para registrar eventos
	PrintTree          bool          `json:"printTree" yaml:"printTree" xml:"printTree" toml:"printTree" gorm:"omitempty,printTree"` // Indica se a árvore deve ser impressa
	TreeFileSource     string        `json:"treeFileSource" yaml:"treeFileSource" xml:"treeFileSource" toml:"treeFileSource" gorm:"omitempty,treeFileSource"`
	ComposerTargetPath string        `json:"composerTargetPath" yaml:"composerTargetPath" xml:"composerTargetPath" toml:"composerTargetPath" gorm:"omitempty,composerTargetPath"`
	Entries            []t.FileEntry `json:"entries" yaml:"entries" xml:"entries" toml:"entries" gorm:"omitempty,entries"`                // Lista de entradas de arquivo
	RootID             uuid.UUID     `json:"rootId" yaml:"rootId" xml:"rootId" toml:"rootId" gorm:"type:uuid,default:uuid_generate_v4()"` // ID do diretório raiz
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
		Entries:            make([]t.FileEntry, 0),
		RootID:             uuid.Nil, // Inicializa o ID do diretório raiz como vazio
		Logger:             logger,
	}

	if err := fte.ParseTree(); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to parse tree source: %s", err.Error()))
		return nil, fmt.Errorf("failed to parse tree source: %s", err.Error())
	}

	// Log the number of entries loaded
	gl.Log("debug", fmt.Sprintf("FileTree parsed with %d entries", len(fte.Entries)))

	return fte, nil
}

func (ft *FileTree) AddEntry(entry t.FileEntry) {
	ft.Entries = append(ft.Entries, entry)
	if ft.RootID == uuid.Nil && entry.Type == "directory" {
		ft.RootID = entry.ID // Define o primeiro diretório como raiz
	}
}
func (ft *FileTree) GetEntryByID(id uuid.UUID) *t.FileEntry {
	for _, entry := range ft.Entries {
		if entry.ID == id {
			return &entry
		}
	}
	return nil // Retorna nil se não encontrar
}
func (ft *FileTree) GetChildren(parentID uuid.UUID) []t.FileEntry {
	var children []t.FileEntry
	for _, entry := range ft.Entries {
		if entry.ParentID == parentID {
			children = append(children, entry)
		}
	}
	return children // Retorna a lista de filhos
}
func (ft *FileTree) Sanitize(dirtyData []byte) error {
	// 1: Remove entradas inválidas ou duplicadas
	validEntries := make([]t.FileEntry, 0)
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
	// Read the tree file and populate the entries
	if treeFileSource != "" && !ft.PrintTree {
		file, err := os.Open(treeFileSource)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Failed to open tree file: %s", err))
			return nil
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue // Ignora linhas vazias
			}
			// Parse the line into a FileEntry
			var entry t.FileEntry
			_, err := fmt.Sscanf(line, "%s %s %s %d", &entry.ID, &entry.ParentID, &entry.Type, &entry.Depth)
			if err != nil {
				gl.Log("error", fmt.Sprintf("Failed to parse line '%s': %s", line, err))
				continue // Ignora linhas com erro de formatação
			}
			// Define o nome do arquivo ou diretório
			entry.Name = filepath.Join(composerTargetPath, entry.ID.String())
			// Verifica se o caminho é absoluto
			if !filepath.IsAbs(entry.Name) {
				// Se não for absoluto, converte para absoluto
				absPath, err := filepath.Abs(entry.Name)
				if err != nil {
					gl.Log("error", fmt.Sprintf("Failed to get absolute path for '%s': %s", entry.Name, err))
					continue // Ignora entradas com erro de caminho
				}
				entry.Name = absPath
			}
			// Verifica se o arquivo ou diretório existe
			if _, err := os.Stat(entry.Name); os.IsNotExist(err) {
				gl.Log("error", fmt.Sprintf("File or directory does not exist: %s", entry.Name))
				continue // Ignora entradas que não existem
			}
			// Adiciona a entrada ao slice
			ft.Entries = append(ft.Entries, entry)
		}
		if err := scanner.Err(); err != nil {
			gl.Log("error", fmt.Sprintf("Error reading tree file: %s", err))
			return nil
		}
		gl.Log("debug", fmt.Sprintf("Loaded %d entries from tree file: %s", len(ft.Entries), treeFileSource))
	} else {
		gl.Log("debug", "No tree file provided, initializing empty FileTree")
	}
	return nil
}
