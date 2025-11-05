package annotation

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"sync"

	"github.com/abiosoft/mold"
)

// TemplateManager manages templates using mold for layout inheritance
type TemplateManager struct {
	mold *mold.Mold
	mu   sync.RWMutex
}

// BlockData represents data that can be passed to templates
type BlockData struct {
	Title   string
	Content interface{}
	Data    interface{}
	Blocks  map[string]interface{}
}

// NewTemplateManager creates a new template manager using mold
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		mold: mold.New(),
	}
}

// LoadFromFS loads templates from an embedded filesystem
func (tm *TemplateManager) LoadFromFS(fs embed.FS, layoutPattern, pagePattern string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Load all templates from the filesystem
	// Mold will handle the layout/page relationship automatically

	// Load layouts
	layoutFiles, err := fs.ReadDir("templates/layouts")
	if err == nil {
		for _, file := range layoutFiles {
			if file.IsDir() {
				continue
			}
			name := "layouts/" + file.Name()
			content, err := fs.ReadFile("templates/" + name)
			if err != nil {
				return fmt.Errorf("failed to read layout %s: %w", name, err)
			}
			if err := tm.mold.ParseTemplate(name, string(content)); err != nil {
				return fmt.Errorf("failed to parse layout %s: %w", name, err)
			}
		}
	}

	// Load pages
	pageFiles, err := fs.ReadDir("templates/pages")
	if err == nil {
		for _, file := range pageFiles {
			if file.IsDir() {
				continue
			}
			name := "pages/" + file.Name()
			content, err := fs.ReadFile("templates/" + name)
			if err != nil {
				return fmt.Errorf("failed to read page %s: %w", name, err)
			}
			if err := tm.mold.ParseTemplate(name, string(content)); err != nil {
				return fmt.Errorf("failed to parse page %s: %w", name, err)
			}
		}
	}

	return nil
}

// Render renders a page template (mold will automatically handle layout inheritance)
func (tm *TemplateManager) Render(w io.Writer, pageName string, data interface{}) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.mold.ExecuteTemplate(w, pageName, data)
}

// RenderWithBlocks renders a template with explicit block definitions
func (tm *TemplateManager) RenderWithBlocks(w io.Writer, templateName string, blocks map[string]interface{}) error {
	return tm.Render(w, templateName, blocks)
}

// AddFuncMap adds custom template functions
func (tm *TemplateManager) AddFuncMap(funcMap template.FuncMap) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Convert template.FuncMap to mold.FuncMap
	moldFuncs := make(map[string]interface{})
	for k, v := range funcMap {
		moldFuncs[k] = v
	}
	tm.mold.Funcs(moldFuncs)
}

// ParseTemplate parses and adds a template dynamically
func (tm *TemplateManager) ParseTemplate(name, content string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.mold.ParseTemplate(name, content)
}
