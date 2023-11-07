# go-annotator

Simple tool to annotate image datasets for classification

## Usage
How to install?

- Install Go
- `go run github.com/lucasew/go-annotation --help`

Alternatively
- Clone this repo
- cd to the source of the repo
- go build .
- Run the binary generated

It's a standard CGO-less Go project so the standard tools one would expect in Go, including cross compilation, will work.

**NOTE:** I didn't test it on Windows yet.

## Features
- Ingest a folder of, seemingly, waste files looking for loadable images and organize to a flat folder structure of PNG files.
- YAML configuration for
  - Defining classification stages that you can use to funnel out unrelated data
  - Defining authentication for users. Users and passwords.
  - Defining instructions for each stage with examples from the dataset.

## Stack
- Golang
- HTMX
- Base structure in HTML + page specific data in Markdown