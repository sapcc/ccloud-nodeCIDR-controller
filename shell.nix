# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company
# SPDX-License-Identifier: Apache-2.0

{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  nativeBuildInputs = [
    addlicense
    go-licence-detector
    go_1_24
    golangci-lint
    gotools # goimports
    kubernetes-controller-tools # controller-gen
    setup-envtest
    # keep this line if you use bash
    bashInteractive
  ];
}
