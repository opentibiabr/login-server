package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGodLoginPolicy(t *testing.T) configs.TestGodLoginPolicy {
	t.Helper()
	path := t.TempDir()
	policy, err := configs.NewTestGodLoginPolicy(configs.GlobalConfigs{
		DBConfigs:         configs.DBConfigs{Host: "127.0.0.1", Port: 3306, Name: "fixture_test", User: "fixture_test_login"},
		GameServerConfigs: configs.GameServerConfigs{Port: 27002},
		LoginServerConfigs: configs.LoginServerConfigs{
			Http: configs.HttpLoginConfigs{Port: 28002},
			TestGodLogin: configs.TestGodLoginConfigs{
				Enabled: "true", Environment: "test", DBName: "fixture_test", DBUser: "fixture_test_login",
				DBHost: "127.0.0.1", DBPort: "3306",
				ServerPath: path, GamePort: "27002", HTTPPort: "28002",
			},
		},
	}, path)
	require.NoError(t, err)
	return policy
}

func TestTestGodLoginSecondFactorBypassesOnlyExactGod(t *testing.T) {
	policy := testGodLoginPolicy(t)
	for _, accountType := range []uint32{1, 4, 5, 6, 7} {
		for _, enrolled := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/enrolled-%t", accountType, enrolled), func(t *testing.T) {
				db, mock := newAuthenticatorMock(t)
				if accountType != accountTypeGod {
					query := mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42))
					if enrolled {
						query.WillReturnRows(authenticatorRow("v1:unused", nil, 0))
					} else {
						query.WillReturnError(sql.ErrNoRows)
					}
				}
				verification, err := (&Account{ID: 42, Type: accountType}).VerifyLoginSecondFactorWithPolicy(
					context.Background(), db, "", "", "", policy)
				switch {
				case accountType == accountTypeGod || (!enrolled && accountType < accountTypeGameMaster):
					assert.NoError(t, err)
				case enrolled:
					assertPublicError(t, err, serviceerrors.CodeAuthenticatorRequired, "AUTHENTICATOR_REQUIRED")
				default:
					assertPublicError(t, err, serviceerrors.CodeStaffAuthenticatorRequired, "STAFF_AUTHENTICATOR_REQUIRED")
				}
				assert.Equal(t, SecondFactorVerification{}, verification)
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	}
}

func TestTestGodLoginDefaultPolicyKeepsGodStrict(t *testing.T) {
	for _, enrolled := range []bool{false, true} {
		db, mock := newAuthenticatorMock(t)
		query := mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42))
		if enrolled {
			query.WillReturnRows(authenticatorRow("v1:unused", nil, 0))
		} else {
			query.WillReturnError(sql.ErrNoRows)
		}
		_, err := (&Account{ID: 42, Type: accountTypeGod}).VerifyLoginSecondFactor(context.Background(), db, "", "", "")
		if enrolled {
			assertPublicError(t, err, serviceerrors.CodeAuthenticatorRequired, "AUTHENTICATOR_REQUIRED")
		} else {
			assertPublicError(t, err, serviceerrors.CodeStaffAuthenticatorRequired, "STAFF_AUTHENTICATOR_REQUIRED")
		}
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestTestGodLoginPolicyDoesNotManufactureSecondFactorProof(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	verification, err := (&Account{ID: 42, Type: accountTypeGod}).VerifyLoginSecondFactorWithPolicy(
		context.Background(), db, "invalid-totp", "invalid-trusted-device", "", testGodLoginPolicy(t))
	require.NoError(t, err)
	assert.Equal(t, SecondFactorVerification{}, verification)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTestGodLoginPolicyPreservesNilInputs(t *testing.T) {
	policy := testGodLoginPolicy(t)
	db, _ := newAuthenticatorMock(t)
	for _, test := range []struct {
		account *Account
		db      *sql.DB
	}{
		{nil, db}, {&Account{ID: 42, Type: accountTypeGod}, nil},
	} {
		_, err := test.account.VerifyLoginSecondFactorWithPolicy(context.Background(), test.db, "", "", "", policy)
		assertPublicError(t, err, serviceerrors.CodeAuthenticatorDataUnavailable, "AUTHENTICATOR_DATA_UNAVAILABLE")
	}
}
