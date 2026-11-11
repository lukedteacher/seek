{ pkgs, lib, config, inputs, ... }:

{
  env.GREET = "devenv";

  packages = [ pkgs.git ];

  languages.go.enable = true;

  scripts.hello.exec = ''
    echo hello from $GREET
  '';

  # Convenience wrapper: `dev -v` shows watcher output, `dev` hides it.
  scripts.dev.exec = ''
    if [ "$1" = "-v" ] || [ "$1" = "--verbose" ]; then
      export DEVENV_PROCESS_VERBOSE=1
    fi
    devenv up
  '';

  enterShell = ''
    hello
    git --version
  '';

  tasks = {
    "templ:generate".exec = "go tool templ generate";
    "templ:generate".execIfModified = [ "**/*.templ" ];

    "css:generate".exec = "go tool gotailwind -i src/input.css -o static/styles/style.css";
    "css:generate".execIfModified = [
      "src/**/*.css"
      "src/daisyui/**/*.js"
      "internal/**/*.templ"
      "internal/**/*.go"
      "pkg/**/*.templ"
    ];

    "sqlc:generate".exec = ''
      GOBIN="$PWD/.bin" go install github.com/delaneyj/toolbelt/sqlc-gen-zombiezen
      PATH="$PWD/.bin/sqlc generate"
    '';
    "sqlc:generate".execIfModified = [
      "sqlc.yaml"
      "queries/**/*.sql"
      "migrations/**/*.sql"
    ];

    "seek:build".exec = "go build -o .tmp/bin/main.exe ./cmd/server";
    "seek:build".after = [ "sqlc:generate" "templ:generate" "css:generate" ];

    "seek:build-linux".exec = "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags=prod -o bin/main ./cmd/server";
    "seek:build-linux".after = [ "sqlc:generate" "templ:generate" "css:generate" ];

    "seek:migrate".exec = "go run ./cmd/server -migrate-only";
    "seek:seed".exec = "go run ./cmd/server -seed-only";
    "seek:reset".exec = "go run ./cmd/server -reset-migrations";
  };

  processes = {
    templ-watch.exec = ''
      if [ -n "$DEVENV_PROCESS_VERBOSE" ]; then
        go tool templ generate -watch
      else
        go tool templ generate -watch >/dev/null 2>&1
      fi
    '';

    css-watch.exec = ''
      if [ -n "$DEVENV_PROCESS_VERBOSE" ]; then
        go tool gotailwind -i src/input.css -o static/styles/style.css -w
      else
        go tool gotailwind -i src/input.css -o static/styles/style.css -w >/dev/null 2>&1
      fi
    '';

    server-watch.exec = "go tool air -c .air.toml";
  };

  enterTest = ''
    echo "Running tests"
    git --version | grep --color=auto "${pkgs.git.version}"
  '';
}
