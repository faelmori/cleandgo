package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	it "github.com/faelmori/cleandgo/interfaces"
	gl "github.com/faelmori/cleandgo/logger"
)

func SanitizeLineIcons(line string, directoriesIcons, filesIcons []string) string {
	unwantedChars := []string{}

	unwantedChars = append(unwantedChars, directoriesIcons...)
	unwantedChars = append(unwantedChars, filesIcons...)

	for _, char := range unwantedChars {
		line = strings.ReplaceAll(line, char, "")
	}
	return strings.TrimSpace(line) // Remove espaços extras ao redor
}
func SanitizeLineChars(line string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_\/.-]`)
	return re.ReplaceAllString(line, "")
}
func SetTreeViewEntriesDeepness(ft it.IFileTree) error {
	var maxDepth int
	depthRegex := regexp.MustCompile(`^([\s│├└──]*)`)
	for i, entry := range ft.GetEntries() {
		matches := depthRegex.FindStringSubmatch(entry.GetOriginName())
		if len(matches) > 0 {
			depth := strings.Count(matches[1], "│") + strings.Count(matches[1], "├") + strings.Count(matches[1], "└")
			ft.GetEntries()[i].SetDepth(depth)
			if depth > maxDepth {
				maxDepth = depth
			}
		} else {
			ft.GetEntries()[i].SetDepth(0)
		}
	}

	// Define a profundidade máxima para o FileTree
	ft.SetMaxDepth(maxDepth)

	gl.Log("debug", fmt.Sprintf("FileTree entries deepness set with %d entries", len(ft.GetEntries())))

	// Define os IDs de cada entrada com base na posição na lista
	if err := SetTreeViewDrawedIdentifiers(ft); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to set tree view identifiers: %s", err))
		return fmt.Errorf("failed to set tree view identifiers: %s", err)
	}

	// Define as referências de estrutura para cada entrada
	if err := SetTreeStructureReferences(ft); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to set tree structure references: %s", err))
		return fmt.Errorf("failed to set tree structure references: %s", err)
	}

	gl.Log("debug", "FileTree structure references set successfully")

	return nil
}
func SetTreeViewDrawedIdentifiers(ft it.IFileTree) error {
	// Define os IDs de cada entrada com base na posição na lista
	for i := range ft.GetEntries() {
		ft.GetEntries()[i].SetID(uuid.New()) // Gera um novo UUID para cada entrada
	}

	// Define o ID do diretório raiz, se ainda não estiver definido
	if ft.GetRootID() == uuid.Nil {
		for _, entry := range ft.GetEntries() {
			if entry.GetType() == "directory" {
				ft.SetRootID(entry.GetID())
				break
			}
		}
	}

	gl.Log("debug", fmt.Sprintf("Root ID set to: %s", ft.GetRootID()))

	if ft.GetRootID() == uuid.Nil {
		gl.Log("error", "No root directory found in FileTree entries")
		return fmt.Errorf("no root directory found in FileTree entries")
	}

	gl.Log("debug", fmt.Sprintf("Set tree view drawed identifiers with %d entries", len(ft.GetEntries())))

	return nil
}
func SetTreeStructureReferences(ft it.IFileTree) error {
	// Cria um mapa para facilitar a busca de entradas por ID
	entryMap := make(map[uuid.UUID]it.IFileEntry)

	for i := range ft.GetEntries() {
		entryMap[ft.GetEntries()[i].GetID()] = ft.GetEntries()[i]
	}

	// Define as referências de estrutura para cada entrada
	for i := range ft.GetEntries() {
		entry := ft.GetEntries()[i]
		if entry.GetParentID() != uuid.Nil {
			if parent, ok := entryMap[entry.GetParentID()]; ok {
				entry.SetParent(parent)
			}
		}
	}
	return nil
}
func ExtractComment(line string) (string, string) {
	re := regexp.MustCompile(`(.*?)\s*#\s*(.*)`) // Captura o texto antes do '#' e o comentário
	matches := re.FindStringSubmatch(line)

	if len(matches) > 2 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]) // Retorna nome limpo e comentário separado
	}
	return line, "" // Se não houver comentário, retorna a linha original
}
func RemoveDrawedIdentifiers(line string, drawedMap map[string]string) string {
	if drawedMap == nil {
		gl.Log("error", "DrawedMap is nil, cannot parse line")
		return ""
	}
	// Substitui os caracteres de estrutura da árvore pelos símbolos correspondentes
	for char, symbol := range drawedMap {
		// Remove os símbolos de estrutura da árvore do nome
		line = strings.ReplaceAll(line, char, symbol)
	}
	return strings.TrimSpace(line) // Retorna a linha sem espaços extras
}
func ContainsIcon(line string, icons []string) bool {
	for _, icon := range icons {
		if strings.Contains(line, icon) {
			return true
		}
	}
	return false
}
