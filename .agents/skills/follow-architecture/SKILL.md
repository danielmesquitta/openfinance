---
name: follow-architecture
description: Write and review Go code for the repository while preserving its package boundaries, dependency flow, adapter patterns, generated files, and test conventions. Use when creating, changing, refactoring, or reviewing code in this repository, especially when deciding whether code belongs in internal/domain, internal/provider, internal/pkg, internal/config, internal/app, or cmd, or when introducing providers, configuration, dependency wiring, mocks, or tests.
---

# Follow Architecture

## Start from the repository

1. Inspect nearby packages and analogous implementations before adding an abstraction.
2. Classify each change as business logic, an external boundary, a reusable helper, system configuration, application composition, or an executable entrypoint.
3. Preserve established package naming, constructors, error wrapping, dependency injection, generated mocks, and test conventions.
4. Keep changes narrowly scoped. Do not restructure unrelated packages.

## Place code by responsibility

### Put business behavior in `internal/domain`

- Put business entities and value types in `internal/domain/entity`.
- Put workflows and application-independent business operations in `internal/domain/usecase/<usecase>`.
- Put reusable structural constraints such as required values, collection sizes, allowed primitive values, and simple field comparisons in `validate` tags on the input or entity they describe.
- Inject `*validator.Validator` into use cases that receive externally constructed input.
- Validate at the start of each public use-case operation, before invoking providers or performing business work.
- Keep semantic invariants, configured-value relationships, normalization, classification, filtering, deduplication, fallback selection, and other business decisions in domain constructors or explicit domain functions.
- Let domain use cases depend on provider interfaces and provider-neutral contracts when they need external capabilities.
- Never import a concrete provider implementation into the domain.
- Keep HTTP payloads, SDK types, environment parsing, database schemas, vendor errors, and transport details out of the domain.

Follow this shortened pattern from `internal/domain/usecase/ingest`:

```go
type IngestInput struct {
	StartDate time.Time `validate:"required"`
	EndDate   time.Time `validate:"required,gtefield=StartDate"`
}

type Ingest struct {
	val *validator.Validator
}

func (s *Ingest) Execute(ctx context.Context, input IngestInput) error {
	if err := s.val.Validate(input); err != nil {
		return fmt.Errorf("invalid ingest input: %w", err)
	}

	// Continue only after validation succeeds.
	return nil
}
```

Inject the validator through the use-case constructor; omit only unrelated constructor dependencies from shortened examples.

### Put external boundaries in `internal/provider`

- Create one root package for each external capability, following packages such as `sheet`, `gpt`, `companyapi`, and `openfinance`.
- Define provider interfaces and provider-neutral request, response, option, and boundary types in the root provider package.
- Put each concrete technology or vendor implementation in a subpackage such as `notionapi`, `openai`, `brasilapi`, or `pluggyapi`.
- Keep transport DTOs private to the concrete adapter. Translate them into domain entities or provider-neutral types at the boundary.
- Pass `context.Context` through external operations.
- Wrap failures with concise, operation-specific context.
- Add a compile-time assertion that each concrete client implements its provider interface.
- Keep authentication, retries, pagination, serialization, SDK use, and transport mapping inside the adapter.
- Keep categorization, financial policy, fallback selection, and every other business decision out of adapters.
- Generate mocks from provider interfaces. Never hand-edit generated mock packages.

Separate the provider port from private adapter transport types:

```go
// internal/provider/openfinance
type APIProvider interface {
	ListTransactionsByIngestProfileID(
		ctx context.Context,
		ingestProfileID string,
		from, to time.Time,
	) ([]entity.Transaction, error)
}

// internal/provider/openfinance/pluggyapi
type listTransactionsResponse struct {
	TotalPages int64                            `json:"totalPages"`
	Results    []listTransactionsResponseResult `json:"results"`
}

var _ openfinance.APIProvider = (*Client)(nil)
```

### Put focused helpers in `internal/pkg`

- Add only reusable, domain-agnostic helpers that do not represent business policy or an external integration.
- Prefer small cohesive packages such as `docutil` or `validator`.
- Keep business-specific helpers in `internal/domain`, even when the helper is small.
- Do not use `internal/pkg` as a miscellaneous dumping ground.
- Do not import `internal/app`, concrete providers, or runtime configuration from helper packages.
- Treat `internal/pkg/validator` as a reusable validation mechanism and error-translation helper.
- Keep the validator unaware of OpenFinance-specific policy.
- Do not move business validation into the generic validator merely because multiple domain types use it.
- Add validation-engine behavior there only when it is independent of a use case, provider, and configuration format.

Keep helpers generic and focused, as in `internal/pkg/docutil`:

```go
func CleanDocument(doc string) string {
	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(doc, "")
}
```

Keep category selection, fallback behavior, and other business-specific functions in `internal/domain` instead.

### Put code-backed system configuration in `internal/config`

- Load environment variables and configuration files here.
- Parse raw environment and configuration-file data here.
- Run structural validation through the shared validator after parsing.
- Use `dive` on configuration wrappers to validate nested entities through their own tags.
- Construct typed domain settings only after structural validation succeeds.
- Let domain constructors perform semantic validation and normalization, and keep them independently safe for callers outside configuration loading.
- Configure logging, timezones, and process-wide runtime behavior here.
- Keep secrets and provider credentials here until `internal/app` passes them to concrete adapters.
- Keep business defaults and business policy in domain constructors rather than configuration loaders.
- Distinguish `internal/config`, which contains Go code, from root `config/`, which contains configuration data and examples.

Validate nested configuration shapes through their wrapper:

```go
type IngestProfilesFileData struct {
	IngestProfiles []entity.IngestProfile `validate:"required,min=1,unique=ID,dive"`
}
```

Validate structure first, then delegate semantic validation and normalization to the domain constructor:

```go
if err := e.val.Validate(e.IngestProfilesFileData); err != nil {
	return fmt.Errorf("failed to validate ingest profiles file: %w", err)
}

settings, err := entity.NewIngestSettings(e.IngestProfiles)
if err != nil {
	return fmt.Errorf("invalid domain settings: %w", err)
}

e.IngestSettings = settings
```

### Compose the system in `internal/app`

- Assemble concrete providers and domain use cases.
- Maintain Wire bindings and constructor extraction helpers in `internal/app/wire.go`.
- Implement CLI and Lambda delivery behavior in their existing `internal/app` subpackages.
- Translate runtime input into use-case input and use-case results into runtime responses.
- Keep business decisions out of handlers and commands.

Bind ports to concrete adapters only in the composition layer:

```go
wire.Build(
	validator.NewValidator,
	config.NewEnv,
	wire.Bind(new(openfinance.APIProvider), new(*pluggyapi.Client)),
	pluggyapi.NewClient,
	ingest.NewIngest,
)
```

Let Wire supply the shared validator to configuration and use-case constructors; do not instantiate it inside either consumer.

### Keep `cmd` minimal

- Select and invoke the corresponding `internal/app` adapter.
- Handle only top-level process termination or logging.
- Do not construct the dependency graph or implement business logic in `cmd`.

Keep the executable entrypoint as small as the existing CLI entrypoint:

```go
func main() {
	if err := cli.Execute(); err != nil {
		log.Fatalf("failed to execute cli: %v", err)
	}
}
```

## Preserve dependency direction

Use the existing dependency flow:

```text
cmd
  -> internal/app
    -> internal/config
    -> internal/domain
    -> internal/provider interfaces and implementations
    -> internal/pkg

internal/domain/usecase
  -> internal/domain/entity
  -> internal/provider root contracts
  -> internal/pkg

internal/provider/<implementation>
  -> its provider root contract
  -> internal/domain/entity when mapping domain data
  -> internal/config when construction requires runtime credentials
```

Never:

- Import concrete provider implementations from domain code.
- Make one adapter call another adapter to decide business behavior.
- Make `internal/pkg` depend on application or adapter packages.
- Bypass `internal/app` from `cmd`.
- Manually edit `internal/app/wire_gen.go` or generated `mock*` packages.

## Implement features in order

1. Model or update domain entities and public use-case inputs.
2. Express generic shape constraints with `validate` tags.
3. Implement semantic invariants and normalization in domain constructors or explicit domain functions.
4. Inject and invoke the shared validator at each input boundary before external or stateful work.
5. Implement business orchestration in the appropriate domain use case.
6. Define or extend a provider root interface only when the feature needs an external capability.
7. Implement the vendor-specific adapter and its transport mappings.
8. Add code-backed runtime settings under `internal/config` when required.
9. Wire dependencies in `internal/app/wire.go`.
10. Keep CLI or Lambda input and output translation in `internal/app`.
11. Run `make generate` after changing provider interfaces, mocks, or Wire bindings.
12. Add tests at the layer that owns the behavior, then run the repository validation commands appropriate to the change.

## Follow project coding conventions

- Accept `context.Context` as the first parameter for cancellable or external operations.
- Name constructors `New<Type>`.
- Apply declarative tags for structural constraints supported by the shared validator.
- Validate public use-case input before any provider call or stateful operation.
- Avoid duplicate hand-written `Validate()` methods when tags clearly express the entire structural rule.
- Keep business-semantic validation explicit in domain constructors or functions when tags cannot clearly express the rule.
- Wrap errors with `%w` and concise operation context.
- Keep external response structures unexported unless they intentionally form a provider-neutral contract.
- Use compile-time interface assertions for concrete providers.
- Co-locate tests with their packages in `_test.go` files.
- Prefer table-driven tests for validation and mapping variants.
- Use generated testify mocks for use-case dependencies.
- Never edit a file marked as generated.

## Use these placement examples

- Put transaction categorization, filtering, or fallback behavior in `internal/domain`.
- Put a new banking, payment, database, or API integration behind a root provider contract and a concrete `internal/provider/<capability>/<vendor>` adapter.
- Put CNPJ formatting in a focused `internal/pkg` helper only when it is reusable and domain-agnostic.
- Load a new environment variable or coded runtime setting in `internal/config`, then inject it through `internal/app`.
- Add a new CLI flag in `internal/app/cli` and translate it into domain use-case input.
- Change Wire bindings in `internal/app/wire.go`, then regenerate `internal/app/wire_gen.go`.

## Test and validate

- Test domain rules directly in domain tests without concrete external clients.
- Test adapter mapping, pagination, validation, and failure handling at the provider implementation layer.
- Test CLI and Lambda input/output translation in `internal/app` with generated use-case mocks.
- Test validation tags directly with `validator.NewValidator().Validate`.
- Cover valid boundary cases as well as missing and invalid fields.
- Test each public use-case operation separately to prove invalid input is rejected before provider calls.
- Test configuration structural failures independently from domain-construction failures.
- Test domain constructors directly for semantic invalidity, fallback behavior, and normalization.

Use compact table-driven tests for domain rules:

```go
func TestShouldIgnoreTransaction(t *testing.T) {
	tests := []struct {
		name        string
		transaction entity.Transaction
		want        bool
	}{
		{name: "incoming credit", transaction: entity.Transaction{
			Direction: entity.TransactionDirectionCredit,
		}, want: true},
		{name: "expense", transaction: entity.Transaction{Name: "Market"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := shouldIgnoreTransaction(test.transaction, false)
			if got != test.want {
				t.Fatalf("shouldIgnoreTransaction() = %t, want %t", got, test.want)
			}
		})
	}
}
```

Use local HTTP servers to test concrete adapters without live external calls:

```go
func TestCreateChatCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writeCompletion(writer, "completed")
	}))
	t.Cleanup(server.Close)

	content, err := newTestClient(server.URL).CreateChatCompletion(t.Context(), "hello")
	if err != nil {
		t.Fatalf("CreateChatCompletion() error = %v", err)
	}
	if content != "completed" {
		t.Fatalf("content = %q", content)
	}
}
```

- Run `make generate` when interfaces or dependency wiring change.
- Run `make test` for the race-enabled repository test suite.
- Run `make build` when changing application composition or executable paths.
- Run `make lint` after code changes and inspect any automatic fixes.
- Run `git diff --check` before handing off the change.
