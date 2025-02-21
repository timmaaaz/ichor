// Package auth provides authentication and authorization support.
// Authentication: You are who you say you are.
// Authorization:  You have permission to do what you are requesting to do.
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/open-policy-agent/opa/rego"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus/stores/permissionscache"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus/stores/permissionsdb"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus/stores/approvaldb"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ErrForbidden is returned when a auth issue is identified.
var ErrForbidden = errors.New("attempted action is not allowed")

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}

// KeyLookup declares a method set of behavior for looking up
// private and public keys for JWT use. The return could be a
// PEM encoded string or a JWS based key.
type KeyLookup interface {
	PrivateKey(kid string) (key string, err error)
	PublicKey(kid string) (key string, err error)
}

// Config represents information required to initialize auth.
type Config struct {
	Log       *logger.Logger
	DB        *sqlx.DB
	KeyLookup KeyLookup
	Issuer    string
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	keyLookup      KeyLookup
	userBus        *userbus.Business
	permissionsBus *permissionsbus.Business
	method         jwt.SigningMethod
	parser         *jwt.Parser
	issuer         string
}

// New creates an Auth to support authentication/authorization.
func New(cfg Config) (*Auth, error) {

	// If a database connection is not provided, we won't perform the
	// user enabled check.
	var userBus *userbus.Business
	var userApprovalStatusBus *approvalbus.Business
	if cfg.DB != nil {
		userApprovalStatusBus = approvalbus.NewBusiness(cfg.Log, nil, approvaldb.NewStore(cfg.Log, cfg.DB))
		userBus = userbus.NewBusiness(cfg.Log, nil, userApprovalStatusBus, usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB), 10*time.Minute))
	}

	var permissionsBus *permissionsbus.Business
	if cfg.DB != nil {
		permissionsBus = permissionsbus.NewBusiness(cfg.Log, permissionscache.NewStore(cfg.Log, permissionsdb.NewStore(cfg.Log, cfg.DB), 24*time.Hour))
	}

	a := Auth{
		keyLookup:      cfg.KeyLookup,
		userBus:        userBus,
		permissionsBus: permissionsBus,
		method:         jwt.GetSigningMethod(jwt.SigningMethodRS256.Name),
		parser:         jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name})),
		issuer:         cfg.Issuer,
	}

	return &a, nil
}

// Issuer provides the configured issuer used to authenticate tokens.
func (a *Auth) Issuer() string {
	return a.issuer
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(kid string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = kid

	privateKeyPEM, err := a.keyLookup.PrivateKey(kid)
	if err != nil {
		return "", fmt.Errorf("private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("parsing private pem: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}

// Authenticate processes the token to validate the sender's token is valid.
func (a *Auth) Authenticate(ctx context.Context, bearerToken string) (Claims, error) {
	if !strings.HasPrefix(bearerToken, "Bearer ") {
		return Claims{}, errors.New("expected authorization header format: Bearer <token>")
	}

	jwt := bearerToken[7:]

	var claims Claims
	token, _, err := a.parser.ParseUnverified(jwt, &claims)
	if err != nil {
		return Claims{}, fmt.Errorf("error parsing token: %w", err)
	}

	kidRaw, exists := token.Header["kid"]
	if !exists {
		return Claims{}, fmt.Errorf("kid missing from header: %w", err)
	}

	kid, ok := kidRaw.(string)
	if !ok {
		return Claims{}, fmt.Errorf("kid malformed: %w", err)
	}

	pem, err := a.keyLookup.PublicKey(kid)
	if err != nil {
		return Claims{}, fmt.Errorf("failed to fetch public key: %w", err)
	}

	input := map[string]any{
		"Key":   pem,
		"Token": jwt,
		"ISS":   a.issuer,
	}

	if err := a.opaPolicyEvaluation(ctx, regoAuthentication, RuleAuthenticate, input); err != nil {
		return Claims{}, fmt.Errorf("authentication failed : %w", err)
	}

	// Check the database for this user to verify they are still enabled.

	if err := a.isUserEnabled(ctx, claims); err != nil {
		return Claims{}, fmt.Errorf("user not enabled : %w", err)
	}

	return claims, nil
}

// Authorize attempts to authorize the user with the provided input roles, if
// none of the input roles are within the user's claims, we return an error
// otherwise the user is authorized.
func (a *Auth) Authorize(ctx context.Context, claims Claims, userID uuid.UUID, rule string, tableInfo TableInfo) error {
	input := map[string]any{
		"Roles":   claims.Roles,
		"Subject": claims.Subject,
		"UserID":  userID,
	}

	if err := a.opaPolicyEvaluation(ctx, regoAuthorization, rule, input); err != nil {
		return fmt.Errorf("rego evaluation failed : %w", err)
	}

	// Authorize on our permissions
	perms, err := a.permissionsBus.QueryUserPermissions(ctx, userID)
	if err != nil {
		return fmt.Errorf("query user permissions: %w", err)
	}

	var zeroValue TableInfo

	// If we have table information in the context, check table permissions
	if tableInfo != zeroValue {
		if !hasTablePermission(perms, tableInfo) {
			return fmt.Errorf("user %s lacks permission for %s on table %s",
				userID, tableInfo.Action, tableInfo.Name)
		}
	}
	return nil
}

// hasTablePermission checks if the user has the required permission for the specified table
func hasTablePermission(userPerms permissionsbus.UserPermissions, tableInfo TableInfo) bool {
	// Search through all roles assigned to the user
	for _, role := range userPerms.Roles {
		// Check each table access in this role
		for _, tableAccess := range role.Tables {
			if strings.EqualFold(tableAccess.TableName, tableInfo.Name) {
				// Check specific permission based on the action
				switch tableInfo.Action {
				case Actions.Create:
					if tableAccess.CanCreate {
						return true
					}
				case Actions.Read:
					if tableAccess.CanRead {
						return true
					}
				case Actions.Update:
					if tableAccess.CanUpdate {
						return true
					}
				case Actions.Delete:
					if tableAccess.CanDelete {
						return true
					}
				}
			}
		}
	}
	return false
}

// opaPolicyEvaluation asks opa to evaluate the token against the specified token
// policy and public key.
func (a *Auth) opaPolicyEvaluation(ctx context.Context, regoScript string, rule string, input any) error {
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", regoScript),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}

// isUserEnabled hits the database and checks the user is not disabled. If the
// no database connection was provided, this check is skipped.
func (a *Auth) isUserEnabled(ctx context.Context, claims Claims) error {
	if a.userBus == nil {
		return nil
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return fmt.Errorf("parse user: %w", err)
	}

	usr, err := a.userBus.QueryByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("query user: %w", err)
	}

	if !usr.Enabled {
		return fmt.Errorf("user disabled")
	}

	return nil
}
