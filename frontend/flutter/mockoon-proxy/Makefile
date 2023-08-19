
install-mockoon-cli:
	@if which mockoon-cli >/dev/null; \
	then echo "mockoon cli already installed!"; \
	else \
		if ! which npm >/dev/null; \
			then echo "npm not found - install nodejs"; \
			else npm install -g @mockoon/cli; \
		fi \
	fi

run:
	dart run lib/server/server.dart

.PHONY: install-mockoon-cli run