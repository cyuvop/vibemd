.PHONY: test mac windows all clean

WAILS := $(HOME)/go/bin/wails
DIST  := dist

test:
	go test ./... -timeout 30s

mac: test
	mkdir -p $(DIST)
	$(WAILS) build -platform darwin/universal -clean
	@if command -v create-dmg >/dev/null 2>&1; then \
		create-dmg \
			--volname "vibemd" \
			--window-size 540 380 \
			--icon-size 128 \
			--app-drop-link 380 190 \
			"$(DIST)/vibemd-mac.dmg" \
			"build/bin/vibemd.app"; \
		echo "→ $(DIST)/vibemd-mac.dmg"; \
	else \
		cp -r build/bin/vibemd.app $(DIST)/vibemd.app; \
		echo "→ $(DIST)/vibemd.app (install create-dmg for DMG packaging)"; \
	fi

windows: test
	mkdir -p $(DIST)
	$(WAILS) build -platform windows/amd64 -clean -nsis
	@echo "→ $(DIST)/vibemd-setup.exe"

all: mac windows

clean:
	rm -rf build/bin dist
