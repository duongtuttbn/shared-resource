package ginswagger

type SwaggerConfig struct {
	SwaggerAPIHost           string `json:"swagger_api_host"`
	SwaggerHiddenPath        string `json:"swagger_hidden_path"`
	SwaggerBasicAuthUser     string `json:"swagger_basic_auth_user"`
	SwaggerBasicAuthPassword string `json:"swagger_basic_auth_password"`
}

type Data struct {
	JSON string `json:"json"`
}
