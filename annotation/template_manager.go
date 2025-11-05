package annotation

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"sync"
)

// TemplateManager manages layouts and pages with block support
type TemplateManager struct {
	layouts   map[string]*template.Template
	pages     map[string]*template.Template
	functions template.FuncMap
	mu        sync.RWMutex
}

// BlockData represents data that can be passed to blocks
type BlockData struct {
	Title   string
	Content interface{}
	Data    interface{}
	Blocks  map[string]interface{}
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		layouts: make(map[string]*template.Template),
		pages:   make(map[string]*template.Template),
		functions: template.FuncMap{
			"block": func(name string, data interface{}) string {
				return fmt.Sprintf("{{/* block:%s */}}", name)
			},
		},
	}
	return tm
}

// LoadFromFS loads templates from an embedded filesystem
func (tm *TemplateManager) LoadFromFS(fs embed.FS, layoutPattern, pagePattern string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Load layouts
	layoutFiles, err := fs.ReadDir("templates/layouts")
	if err == nil {
		for _, file := range layoutFiles {
			if file.IsDir() {
				continue
			}
			name := file.Name()
			tmpl, err := template.New(name).Funcs(tm.functions).ParseFS(fs, "templates/layouts/"+name)
			if err != nil {
				return fmt.Errorf("failed to parse layout %s: %w", name, err)
			}
			tm.layouts[name] = tmpl
		}
	}

	// Load pages
	pageFiles, err := fs.ReadDir("templates/pages")
	if err == nil {
		for _, file := range pageFiles {
			if file.IsDir() {
				continue
			}
			name := file.Name()
			tmpl, err := template.New(name).Funcs(tm.functions).ParseFS(fs, "templates/pages/"+name)
			if err != nil {
				return fmt.Errorf("failed to parse page %s: %w", name, err)
			}
			tm.pages[name] = tmpl
		}
	}

	return nil
}

// Render renders a page with a layout
func (tm *TemplateManager) Render(w io.Writer, layoutName, pageName string, data interface{}) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	layout, ok := tm.layouts[layoutName]
	if !ok {
		return fmt.Errorf("layout %s not found", layoutName)
	}

	page, ok := tm.pages[pageName]
	if !ok {
		return fmt.Errorf("page %s not found", pageName)
	}

	// Execute page template first to capture blocks
	var pageBuffer bytes.Buffer
	if err := page.Execute(&pageBuffer, data); err != nil {
		return fmt.Errorf("failed to execute page template: %w", err)
	}

	// Prepare block data
	blockData := &BlockData{
		Data:   data,
		Blocks: make(map[string]interface{}),
	}

	// If data has Title and Content fields, extract them
	if bd, ok := data.(BlockData); ok {
		blockData.Title = bd.Title
		blockData.Content = template.HTML(pageBuffer.String())
		blockData.Data = bd.Data
		for k, v := range bd.Blocks {
			blockData.Blocks[k] = v
		}
	} else if v, ok := data.(map[string]interface{}); ok {
		if title, ok := v["Title"].(string); ok {
			blockData.Title = title
		}
		blockData.Content = template.HTML(pageBuffer.String())
		blockData.Data = data
	} else {
		blockData.Content = template.HTML(pageBuffer.String())
		blockData.Data = data
	}

	// Execute layout with block data
	return layout.Execute(w, blockData)
}

// RenderWithBlocks renders a page with explicit block definitions
func (tm *TemplateManager) RenderWithBlocks(w io.Writer, layoutName string, blocks map[string]interface{}) error {
	tm.mu.RLock()
	layout, ok := tm.layouts[layoutName]
	tm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("layout %s not found", layoutName)
	}

	return layout.Execute(w, blocks)
}

// AddLayout adds a layout template dynamically
func (tm *TemplateManager) AddLayout(name string, tmpl *template.Template) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.layouts[name] = tmpl
}

// AddPage adds a page template dynamically
func (tm *TemplateManager) AddPage(name string, tmpl *template.Template) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.pages[name] = tmpl
}

// ParseLayout parses and adds a layout from a string
func (tm *TemplateManager) ParseLayout(name, content string) error {
	tmpl, err := template.New(name).Funcs(tm.functions).Parse(content)
	if err != nil {
		return err
	}
	tm.AddLayout(name, tmpl)
	return nil
}

// ParsePage parses and adds a page from a string
func (tm *TemplateManager) ParsePage(name, content string) error {
	tmpl, err := template.New(name).Funcs(tm.functions).Parse(content)
	if err != nil {
		return err
	}
	tm.AddPage(name, tmpl)
	return nil
}
