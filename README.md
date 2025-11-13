# go-annotation

> Simple and powerful tool to annotate image datasets for classification

Modern web-based annotation tool with a clean UI, keyboard shortcuts, and collaborative features. Built with Go, HTMX, DaisyUI, and TailwindCSS.

## ✨ Features

- 🎨 **Modern UI** - Beautiful interface with DaisyUI and TailwindCSS
- ⌨️ **Keyboard Shortcuts** - Annotate faster with number keys (1-9) and `?` for unsure
- 🌓 **Dark Mode** - Theme toggle with localStorage persistence
- 🔐 **Authentication** - Multi-user support with password protection
- 📊 **Conditional Tasks** - Create annotation workflows with dependencies
- 🎯 **Task Types** - Boolean, rotation, and custom classification tasks
- 🌍 **i18n Support** - Internationalization for multiple languages
- 📱 **Responsive** - Works on desktop and mobile devices
- 🚀 **Fast** - No CGO dependencies, pure Go with SQLite

## 🚀 Quick Start

### Initialize a Project

```bash
# Create config, database and empty image folder
go-annotation folder

# Ingest a folder of messy files to a images folder
go-annotation ingest ./messy-folder ./images

```

### Start Annotating

```bash
go-annotation folder/config.yaml
```

Then open http://localhost:8080 in your browser!

## ⚙️ Configuration

There is a ready example in ./examples/test for you to play!

### Sample config.yaml

```yaml
meta:
  description: |
    My Image Classification Project
    Instructions for annotators go here.

# Define users and passwords
auth:
  admin:
    password: "secure_password"
  annotator1:
    password: "another_password"

# Define annotation tasks
tasks:
  # Simple classification
  - id: quality
    name: "Image Quality"
    short_name: "Quality"
    classes:
      good:
        name: "Good"
        description: "Clear and well-focused"
      bad:
        name: "Bad"
        description: "Blurry or poor quality"

  # Boolean task (uses built-in type)
  - id: contains_animal
    name: "Contains an animal?"
    type: boolean  # Creates Yes/No automatically

  # Conditional task (only shown based on previous answer)
  - id: animal_type
    name: "What type of animal?"
    if:
      contains_animal: "true"  # Only show if previous is "Yes"
    classes:
      cat:
        name: "Cat"
      dog:
        name: "Dog"
      bird:
        name: "Bird"
      other:
        name: "Other"

  # Rotation detection (uses built-in type)
  - id: rotation
    name: "Image Rotation"
    type: rotation  # OK, ±90°, 180°, H/V flip

# Internationalization (optional)
i18n:
  - name: "Welcome"
    value: "Bem-vindo"
  - name: "Help"
    value: "Ajuda"
```

### Task Types

**Built-in types:**
- `boolean` - Yes/No questions
- `rotation` - Detect image rotation/flipping
- Custom - Define your own classes

**Conditional tasks:**
Use the `if` field to create dependent tasks:
```yaml
- id: second_task
  if:
    first_task: "expected_value"
```

### Authentication

Add users in the `auth` section:
```yaml
auth:
  username:
    password: "plaintext_password"
```

⚠️ **Security Note**: Passwords are stored in plaintext in the config. Use strong passwords and keep your config file secure.

## Architecture

### Stack
- **Backend**: Go templates with HTMX for SPA-like interactions
- **Frontend**: DaisyUI + TailwindCSS with @tailwindcss/typography
- **Templates**: Mold for layout inheritance
- **Database**: SQLite (modernc.org/sqlite - pure Go, no CGO)

### Project Structure
```
go-annotation/
├── annotation/          # Core annotation logic
│   ├── templates/      # Mold templates
│   │   ├── layouts/    # Base layouts
│   │   └── pages/      # Page templates
│   └── assets/         # Generated CSS
├── cmd/                # CLI commands
│   ├── init.go        # Project initialization
│   ├── annotator.go   # Web server
│   ├── ingest.go      # Image import
│   └── query.go       # Database queries
└── examples/          # Sample projects
```

## Development

### Prerequisites
- Mise

See mise.toml for details on commands

### Progress Tracking
The system automatically tracks:
- Completed annotations
- Uncertain annotations (marked with `?`)
- User attribution
- Annotation order

---

**Made with ❤️ using Go, HTMX, DaisyUI and Claude Code**

> The problem is not using AI, it's not setting up the project to be testable and reviewing its outputs
