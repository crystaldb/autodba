# Variables
INSTALL_SCRIPT := ./install.sh
UNINSTALL_SCRIPT := ./uninstall.sh
SYSTEM := $(system)
USER_INSTALL_DIR := $(install_dir)
CONFIG := $(config)

# Install target
install:
	@echo "Running install.sh script..."
	$(INSTALL_SCRIPT) $(if $(SYSTEM), --system) $(if $(USER_INSTALL_DIR), --install-dir $(USER_INSTALL_DIR)) $(if $(CONFIG), --config $(CONFIG))

# Uninstall target
uninstall:
	@echo "Running uninstall.sh script..."
	$(UNINSTALL_SCRIPT) $(if $(SYSTEM), --system) $(if $(USER_INSTALL_DIR), --install-dir $(USER_INSTALL_DIR))

# Clean target
clean:
	rm -rf build_output

.PHONY: install uninstall clean
