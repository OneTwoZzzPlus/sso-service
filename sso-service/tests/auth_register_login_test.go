package tests

import (
	ssov1 "diaryhub/sso-service/protos/gen/auth"
	"diaryhub/sso-service/tests/suite"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID = 0
	appID      = 1
	appSecret  = "test-app-secret"

	passDefaultLen = 10
)

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func Test_RegisterLogin_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParced, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParced.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"], deltaSeconds)
}

func Test_Register_DoubleRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func Test_Register_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name     string
		email    string
		password string
		expError string
	}{
		{
			name:     "Empty password",
			email:    gofakeit.Email(),
			password: "",
			expError: "password is required",
		},
		{
			name:     "Empty email",
			email:    "",
			password: randomFakePassword(),
			expError: "email is required",
		},
		{
			name:     "Empty email and password",
			email:    "",
			password: "",
			expError: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expError)
		})
	}
}

func Test_Login_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	// create user
	email := gofakeit.Email()
	password := randomFakePassword()
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	tests := []struct {
		name     string
		email    string
		password string
		appID    int32
		expError string
	}{
		{
			name:     "Empty password",
			email:    gofakeit.Email(),
			password: "",
			appID:    appID,
			expError: "password is required",
		},
		{
			name:     "Empty email",
			email:    "",
			password: randomFakePassword(),
			appID:    appID,
			expError: "email is required",
		},
		{
			name:     "Empty email and password",
			email:    "",
			password: "",
			appID:    appID,
			expError: "email is required",
		},
		{
			name:     "Empty email and password",
			email:    "",
			password: "",
			appID:    appID,
			expError: "email is required",
		},
		{
			name:     "Non-existent email",
			email:    gofakeit.Email(),
			password: randomFakePassword(),
			appID:    appID,
			expError: "invalid credentials",
		},
		{
			name:     "Non-matching password",
			email:    email,
			password: randomFakePassword(),
			appID:    appID,
			expError: "invalid credentials",
		},
		{
			name:     "Empty app",
			email:    email,
			password: password,
			appID:    emptyAppID,
			expError: "app_id is required",
		},
		{
			name:     "Non-existent app",
			email:    email,
			password: password,
			appID:    gofakeit.Int32(),
			expError: "invalid app id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expError)
		})
	}
}
