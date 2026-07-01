package apiconvention

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// grandfathered lists app-layer response types that already violate the
// unpaged-list envelope convention and predate this guard.
//
// docs/layer-patterns.md prescribes exactly two collection shapes:
//   - paginated  -> query.Result[T]  ({items, total, page, rows_per_page})
//   - unpaged    -> bare-array wrapper (type Entities []Entity)
//
// A struct that exposes json:"items" WITHOUT json:"total" is query.Result[T]
// with its pagination fields amputated and matches neither shape. The entries
// below are known deviations awaiting the frontend-driven reshape (they feed
// dynamic/config-driven UI, so changing them is cross-cutting and must be
// sequenced from the FE). Do NOT add to this list: a NEW violation must be
// fixed to a documented shape, not grandfathered.
//
// Keys are package-qualified ("pkg.Type"), not bare type names: a bare-name
// allowlist would silently exempt an unrelated, newly-introduced type in a
// different package that happens to share one of these names.
var grandfathered = map[string]bool{
	"dataapp.TableConfigList": true, // app/domain/dataapp
}

// Test_ResponseEnvelope_ItemsRequiresTotal enforces that any app-layer response
// type (one that implements Encode) exposing a json:"items" field also exposes
// json:"total". This blocks new totals-less {items} envelopes, which match no
// documented convention (see docs/layer-patterns.md).
func Test_ResponseEnvelope_ItemsRequiresTotal(t *testing.T) {
	found := detectItemsWithoutTotal(t)

	var offenders []string
	for qualifiedName, pos := range found {
		if grandfathered[qualifiedName] {
			continue
		}
		offenders = append(offenders, qualifiedName+" ("+pos+")")
	}
	sort.Strings(offenders)

	if len(offenders) > 0 {
		t.Fatalf("response type(s) expose json:\"items\" without json:\"total\".\n"+
			"Use query.Result[T] if the endpoint is paginated, or a bare-array wrapper "+
			"(type Entities []Entity) if it is not — see docs/layer-patterns.md.\nOffenders:\n  %s",
			strings.Join(offenders, "\n  "))
	}
}

// Test_Grandfathered_StillViolate keeps the allowlist honest: every grandfathered
// type must still be a detected violation. If one no longer is (e.g. it was
// reshaped to a documented convention), this fails so the stale entry gets
// removed. It also proves the detector actually fires.
func Test_Grandfathered_StillViolate(t *testing.T) {
	found := detectItemsWithoutTotal(t)

	for qualifiedName := range grandfathered {
		if _, ok := found[qualifiedName]; !ok {
			t.Errorf("grandfathered type %q is no longer a json:\"items\"-without-\"total\" "+
				"violation — remove it from the grandfathered allowlist", qualifiedName)
		}
	}
}

// detectItemsWithoutTotal scans app/domain for types that implement Encode and
// expose json:"items" but not json:"total". Returns qualified name
// ("pkg.Type") -> source position.
//
// Known limitation (not fixed here — documented follow-up): encoder detection
// via collectEncoders is syntactic — it looks for an `Encode` FuncDecl with a
// matching receiver in the same package. A type that INHERITS Encode from an
// anonymously embedded base (method promotion) is never recognized as an
// encoder and is therefore never scanned by this guard. Resolving that would
// require full method-set resolution (effectively a mini type-checker), which
// is out of scope for a stdlib-only static guard.
func detectItemsWithoutTotal(t *testing.T) map[string]string {
	t.Helper()

	domainDir := filepath.Join(repoRoot(t), "app", "domain")
	out := map[string]string{}

	err := filepath.WalkDir(domainDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		fset := token.NewFileSet()
		pkgs, perr := parser.ParseDir(fset, path, func(fi fs.FileInfo) bool {
			return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if perr != nil {
			// Skip a directory we can't parse rather than failing the whole
			// sweep, but leave a trace — parse errors otherwise surface via
			// `go build` elsewhere, and a silent skip here would mean this
			// guard never actually looked at the directory.
			t.Logf("apiconvention: skipping unparseable dir %s: %v", path, perr)
			return nil
		}

		for _, pkg := range pkgs {
			encoders := collectEncoders(pkg)
			structs := collectStructs(pkg)
			for name, st := range structs {
				if !encoders[name] {
					continue
				}
				if structHasJSONKey(st, "items", structs) && !structHasJSONKey(st, "total", structs) {
					out[pkg.Name+"."+name] = fset.Position(st.Pos()).String()
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking app/domain: %v", err)
	}

	return out
}

// collectEncoders returns the set of type names in pkg that have a method
// `Encode() ([]byte, string, error)` — i.e. web response types.
func collectEncoders(pkg *ast.Package) map[string]bool {
	enc := map[string]bool{}
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "Encode" || fn.Recv == nil {
				continue
			}
			if !isEncodeSig(fn.Type) {
				continue
			}
			if name := receiverName(fn.Recv); name != "" {
				enc[name] = true
			}
		}
	}
	return enc
}

// collectStructs maps struct type names to their definition in pkg.
func collectStructs(pkg *ast.Package) map[string]*ast.StructType {
	out := map[string]*ast.StructType{}
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if st, ok := ts.Type.(*ast.StructType); ok {
					out[ts.Name.Name] = st
				}
			}
		}
	}
	return out
}

// isEncodeSig reports whether ft matches `() ([]byte, string, error)`.
func isEncodeSig(ft *ast.FuncType) bool {
	if ft.Params != nil && ft.Params.NumFields() != 0 {
		return false
	}
	return ft.Results != nil && ft.Results.NumFields() == 3
}

// receiverName extracts the base type name from a method receiver, unwrapping
// pointers (*T) and generics (T[U]).
func receiverName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	expr := recv.List[0].Type
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.IndexExpr:
			expr = t.X
		case *ast.IndexListExpr:
			expr = t.X
		case *ast.Ident:
			return t.Name
		default:
			return ""
		}
	}
}

// structHasJSONKey reports whether st (or a struct it anonymously, untagged
// embeds) has an exported field whose json tag name is key. structs is the
// owning package's struct index (from collectStructs), used to resolve
// embedded types so that e.g. a shared `Pagination{ Total int
// `json:"total"`}` embedded untagged into a response type correctly
// contributes "total" to that type's marshaled JSON.
//
// Only exported fields count: encoding/json omits unexported fields from the
// wire, so an unexported field carrying a json tag (e.g. a lowercase `total`
// typo) must not be treated as satisfying the key.
func structHasJSONKey(st *ast.StructType, key string, structs map[string]*ast.StructType) bool {
	return structHasJSONKeyVisited(st, key, structs, map[string]bool{})
}

// structHasJSONKeyVisited does the work for structHasJSONKey, tracking
// already-visited embedded type names to guard against infinite recursion on
// cyclic embeds.
func structHasJSONKeyVisited(st *ast.StructType, key string, structs map[string]*ast.StructType, visited map[string]bool) bool {
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			// Anonymous field. Only an UNTAGGED embed contributes its own
			// fields to the marshaled JSON under their own keys — a tagged
			// anonymous field (`Foo `json:"foo"``) marshals as a single
			// nested object keyed "foo", not promoted, so it can't satisfy
			// an "items"/"total" top-level key and is correctly skipped.
			if field.Tag != nil {
				continue
			}
			embeddedName := embeddedTypeName(field.Type)
			if embeddedName == "" || visited[embeddedName] {
				continue
			}
			embeddedSt, ok := structs[embeddedName]
			if !ok {
				// Not resolvable via the local package index (e.g. a
				// qualified pkg.Type embed from another package). Skip
				// safely rather than guessing.
				continue
			}
			visited[embeddedName] = true
			if structHasJSONKeyVisited(embeddedSt, key, structs, visited) {
				return true
			}
			continue
		}

		if !field.Names[0].IsExported() {
			continue
		}
		if field.Tag == nil {
			continue
		}
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		jsonTag := tag.Get("json")
		if jsonTag == "" {
			continue
		}
		if strings.Split(jsonTag, ",")[0] == key {
			return true
		}
	}
	return false
}

// embeddedTypeName extracts the base type name from an anonymous field's type
// expression, unwrapping a pointer (*T). A qualified selector (pkg.Type)
// refers to a type from another package that this file's local struct index
// can't resolve, so it returns "" in that case rather than guessing.
func embeddedTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

// repoRoot walks up from the working directory until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate go.mod above %s", dir)
		}
		dir = parent
	}
}
