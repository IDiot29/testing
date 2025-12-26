# My CV using LaTeX and nix

LaTeX CV built with Nix and encrypted secrets management using sops-nix.

## Prerequisites

- Nix with flakes enabled
- age encryption key (stored in `age-key.txt`)

## Building the CV

```bash
# Set the age key
export SOPS_AGE_KEY="$(grep AGE-SECRET-KEY age-key.txt)"

# Build the PDF
nix build --impure

# Output will be in: result/cv-rivaldo-silalahi.pdf
```

## Development

### Enter development shell

```bash
nix develop
```

This provides access to:

- LaTeX environment (latexmk, pdflatex, etc.)
- sops for secrets management
- age for encryption

### Edit the CV

1. Edit `cv.tex` with your changes
2. Build to test your changes

### Managing secrets

The phone number is encrypted in `secrets.yaml`.

To view encrypted secrets:

```bash
export SOPS_AGE_KEY_FILE=$PWD/age-key.txt
nix develop -c sops secrets.yaml
```

To edit encrypted secrets:

```bash
export SOPS_AGE_KEY_FILE=$PWD/age-key.txt
nix develop -c sops secrets.yaml
```

This will open your editor. Save and quit to re-encrypt.

## Project structure

```
.
├── cv.tex              # CV source (phone number is a placeholder)
├── secrets.yaml        # Encrypted secrets (phone number)
├── age-key.txt         # Private encryption key (DO NOT COMMIT)
├── .sops.yaml          # sops configuration
├── flake.nix           # Nix build configuration
└── flake.lock          # Nix dependencies lock
```

## Important notes

- `secrets.yaml` is safe to commit (it's encrypted)
- The `--impure` flag is required because the build reads environment variables
