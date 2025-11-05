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

### Installation

**Option 1: Install from source**
```bash
go install github.com/lucasew/go-annotation@latest
```

**Option 2: Clone and build**
```bash
git clone https://github.com/lucasew/go-annotation.git
cd go-annotation
go build .
```

### Initialize a Project

```bash
# Create config and database with your images
go-annotation init --images-dir ./my-images

# Or just create config and empty database
go-annotation init
```

This creates:
- `config.yaml` - Sample configuration file
- `annotations.db` - SQLite database for annotations

### Start Annotating

```bash
go-annotation annotator \
  --config config.yaml \
  --database annotations.db \
  --images ./my-images
```

Then open http://localhost:8080 in your browser!

## 📖 Usage Guide

### Commands

#### `init` - Initialize a new project
```bash
go-annotation init [flags]

Flags:
  -i, --images-dir string   Directory containing images to annotate
  -c, --config string       Configuration file to create (default "config.yaml")
  -d, --database string     Database file to create (default "annotations.db")

Examples:
  go-annotation init --images-dir ./photos
  go-annotation init --images-dir ./data/images --config project.yaml
```

#### `annotator` - Start the annotation web server
```bash
go-annotation annotator [flags]

Flags:
  -c, --config string      Config file for the annotation (required)
  -d, --database string    Where to store the annotation database (required)
  -i, --images string      Directory containing images (required)
  -a, --addr string        Server address (default ":8080")

Example:
  go-annotation annotator -c config.yaml -d annotations.db -i ./images
```

#### `ingest` - Import images from nested directories
```bash
go-annotation ingest <source-dir> <destination-dir>

Example:
  go-annotation ingest ./messy-photos ./clean-images
```

Converts nested directory structures into a flat folder of PNG files.

#### `query` - Query the annotation database
```bash
go-annotation query -d annotations.db -t task_id

Example:
  go-annotation query -d annotations.db -t quality
```

## ⚙️ Configuration

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

## ⌨️ Keyboard Shortcuts

- `1-9` - Select classification option 1-9
- `?` - Mark as "Not Sure"
- Mouse click works too!

## 🏗️ Architecture

### Stack
- **Backend**: Go with HTMX for dynamic interactions
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

## 🎨 Development

### Prerequisites
- Go 1.24+
- Node.js 22+ (for TailwindCSS)
- Mise or npm

### Build CSS
```bash
npm install
npm run build:css
```

### Run Tests
```bash
go test ./...
```

### Build
```bash
go build .
```

## 📝 Examples

Check the `examples/` directory for sample projects:
```bash
cd examples/simple-classifier
go-annotation init --images-dir ./images
go-annotation annotator -c config.yaml -d annotations.db -i ./images
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

See LICENSE file for details.

## 🔧 Troubleshooting

**Images not showing up?**
- Ensure images are in a flat directory structure
- Use `ingest` command to flatten nested directories
- Check that images are valid PNG/JPG files

**Can't log in?**
- Check `auth` section in config.yaml
- Verify username and password match

**Database errors?**
- Delete annotations.db and run `init` again
- Check file permissions

## 🚀 Deployment

### Docker
```dockerfile
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN npm install && npm run build:css
RUN CGO_ENABLED=0 go build -o go-annotation

FROM alpine:latest
COPY --from=builder /app/go-annotation /usr/local/bin/
ENTRYPOINT ["go-annotation"]
```

### Systemd Service
```ini
[Unit]
Description=go-annotation server
After=network.target

[Service]
Type=simple
User=annotator
WorkingDirectory=/opt/go-annotation
ExecStart=/usr/local/bin/go-annotation annotator -c config.yaml -d annotations.db -i /data/images
Restart=always

[Install]
WantedBy=multi-user.target
```

## 📊 Exporting Data

Query your annotations:
```bash
# Get all annotations for a task
go-annotation query -d annotations.db -t quality

# Export to CSV (use sqlite3)
sqlite3 annotations.db "SELECT * FROM task_quality" -csv > quality.csv
```

## 🌟 Features in Detail

### Conditional Workflows
Create multi-stage annotation pipelines:
```yaml
tasks:
  - id: has_face
    type: boolean
  - id: face_emotion
    if: { has_face: "true" }
    classes: { happy: {}, sad: {}, neutral: {} }
```

### Example Images in Help
Add example images to help annotators:
```yaml
classes:
  cat:
    name: "Cat"
    examples:
      - "abc123hash"  # Image hash from your dataset
```

### Progress Tracking
The system automatically tracks:
- Completed annotations
- Uncertain annotations (marked with `?`)
- User attribution
- Annotation order

---

**Made with ❤️ using Go, HTMX, and DaisyUI**
