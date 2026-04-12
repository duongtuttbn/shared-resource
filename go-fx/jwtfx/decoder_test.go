package jwtfx

import (
	"encoding/json"
	"testing"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestHMACDecoder(t *testing.T) {
	d := newHMACDecoder([]byte("secret"))
	tokenStr := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	jwtToken, err := d.Decode(tokenStr)
	assert.NoError(t, err)

	assert.Equal(t, "HS256", jwtToken.Header["alg"])
	assert.Equal(t, "JWT", jwtToken.Header["typ"])

	assert.Equal(t, "1234567890", jwtToken.Claims.(jwt.MapClaims)["sub"])
}

func TestRSADecoder(t *testing.T) {
	// Get the JWK as JSON.
	jwksJSON := json.RawMessage(`{"keys":[{"kty":"RSA","e":"AQAB","use":"sig","kid":"4rYKldB778ftdh2WCsw2o5SfIF37V9pRN_ysnFG-Z5I","alg":"RS256","n":"hMiI1-FlzOxAyTHSGiBnmeyoHgyMwHT6tXxlQEJ7c8X1BDzmMAPBGoJuRb33_D5tpNPzB2jqdO1H0YBdifhxMt8Gr4IQBpMYeq5ScFPUFfn6OyQGOjcVdB-_vmLmHyPDZsLoEl8E_J4MPn8rHjXjs6rLUwCpDxeoxmgjqWuP2nIidCEZdPh9pZuq7YxAQjTYduKnBHEbpLc8qVyriqoAphn7JgYUj5Mt8C5Uk8S_fJn5LrZAslJyqC0AONYOl7cqlsNKDpqFYkQ7Dn4FfFCMIZpp6lP4_ZYbZNbJaH5gn0QE0efsMnBDGdOXBkogjWXeASBHv-PoBiqK9L9iBJsj_w"}]}`)
	k, err := keyfunc.NewJWKSetJSON(jwksJSON)
	assert.NoError(t, err)
	d, err := newDecoder(k.KeyfuncCtx)
	assert.NoError(t, err)

	jwtB64 := "eyJhbGciOiJSUzI1NiIsImtpZCI6IjRyWUtsZEI3NzhmdGRoMldDc3cybzVTZklGMzdWOXBSTl95c25GRy1aNUkiLCJ0eXAiOiJKV1QifQ.eyJzdWIiOiIxMjM0NTY3ODkwIn0.JfW4Lm5zRd8LIil4o5GWvBSQslZ5LgwkHSJTh-J9VJ64Hdvu7JOcU30dCCJk26ZKe1xWBDwLaqEvjJch_rcgO7sDQ2oTkp_zU6nB3a7NrFWQCGe3Gw2zu9lI-J6tJRkE42KSA0B9mfGIqotg2bEzmJPCkVn5TkGj6X4_pZM4ZhT3aERv80a9Lw9gvHbdyISclcb-sMsffBw7on_yT289b12EtyLI-AxqcRwS4pKxN3VEw0-_uallOQcqDPmGlYZY1dLdELoA2YzcZ3vOMKxlAZxCOnWLpOuD-nTkrWqkb6iDT3W-85VkzQtJzyB5QP690Aar_Kse_K3ZmHt50fzxzA"

	token, err := d.Decode(jwtB64)
	assert.NoError(t, err)

	assert.Equal(t, "1234567890", token.Claims.(jwt.MapClaims)["sub"])
}

func TestEdDSADecoder(t *testing.T) {
	// Get the JWK as JSON.
	jwkJSON := json.RawMessage(`{"kty":"OKP","d":"hB9_C6VXPe2RhcotUxGMxpujiBawb2c5JL8ItIlWXsk","use":"sig","crv":"Ed25519","kid":"47WxR6uaKt-mPtKtyg-SSEnCVW3AB9VV64dVTHGB5vQ","x":"KsfoYiJ360dya5OcNzRJZ3xwyCCDieE7FlyBXfvh8cc","alg":"EdDSA"}`)
	k, err := keyfunc.NewJWKJSON(jwkJSON)
	assert.NoError(t, err)
	d, err := newDecoder(k.KeyfuncCtx)
	assert.NoError(t, err)

	jwtB64 := "eyJhbGciOiJFZERTQSIsImtpZCI6IjQ3V3hSNnVhS3QtbVB0S3R5Zy1TU0VuQ1ZXM0FCOVZWNjRkVlRIR0I1dlEiLCJ0eXAiOiJKV1QifQ.eyJzdWIiOiIxMjM0NTY3ODkwIn0.bgFynWUILkyyduOhUi1EwIEqsqtqce7LVxK6fg05ID74U67Et7C1ChjNQvan7WeCCKYFqY-cSAnBiIq01-hCAg"
	token, err := d.Decode(jwtB64)
	assert.NoError(t, err)

	assert.Equal(t, "1234567890", token.Claims.(jwt.MapClaims)["sub"])
}
