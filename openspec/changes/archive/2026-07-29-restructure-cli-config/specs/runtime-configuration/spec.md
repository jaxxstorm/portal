## ADDED Requirements

### Requirement: Command-Aware Configuration Resolution
Portal MUST apply the existing configuration precedence to serving commands and
MUST validate only settings relevant to the selected command. OAuth App,
cleanup, and configuration-management commands MUST NOT require a service port
or validate a configured proxy target.

#### Scenario: Serve command uses configuration-file port
- **WHEN** `~/.portal/config.yml` supplies a valid port and a user runs `portal serve` without a positional port
- **THEN** Portal uses the configured port for serving

#### Scenario: Cleanup command ignores invalid serving configuration
- **WHEN** `~/.portal/config.yml` contains an invalid serving setting and a user runs `portal cleanup`
- **THEN** Portal performs cleanup without rejecting the unrelated serving setting

### Requirement: Configuration Template Generation
Portal MUST provide `portal config init` to generate a commented YAML template
at `~/.portal/config.yml`. Portal MUST create the parent directory as needed,
MUST generate a tailnet-private default configuration, and MUST NOT overwrite an
existing configuration file unless `--force` is provided.

#### Scenario: New configuration template is created
- **WHEN** a user runs `portal config init` and the default configuration file does not exist
- **THEN** Portal creates a commented YAML template with Funnel disabled by default
- **AND** Portal reports the generated file location

#### Scenario: Existing configuration is protected
- **WHEN** a user runs `portal config init` and the default configuration file already exists without `--force`
- **THEN** Portal exits with an error and leaves the existing file unchanged

#### Scenario: Explicit replacement is allowed
- **WHEN** a user runs `portal config init --force`
- **THEN** Portal replaces the existing configuration file with the template
