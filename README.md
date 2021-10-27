# Gonaltux

Next generation **A.N.A.L.T.U.X** for Matrix. This is the info bot residing in
`#bsdforen.de:matrix.org` and through the Matrix <-> IRC bridge in `#bsdforen.de` in Libera.

## Build From Source

1. Install Go `>=1.17`
1. Clone the Gonaltux repository: `git clone https://github.com/bsdforen/gonalmatrix`
1. Run `make` from the source directory

## Getting Started

See usage with:

```bash
gonalmatrix-<OS> --help
```

Update config file:

```bash
$EDITOR configs/gonalmatrix.ini
```

Run Gonaltux:

```bash
gonalmatrix-<OS> -c configs/gonalmatrix.ini
```
