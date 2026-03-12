package main

import (
	"fmt"
	"HwWach/internal/app"
	_ "HwWach/docs"
)

//	@title			HwWach API
//	@version		1.0
//	@description	API для управления устройствами и фотографиями

//	@contact.name	API Support
//	@contact.email	support@hwwach.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0

//	@host		localhost:8080
//	@BasePath	/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description			Введите JWT токен в формате "Bearer <token>"

//	@termsOfService	http://swagger.io/terms/

func main() {
	app, err := app.NewApp()
	if err != nil {
		fmt.Println("Error starting app:", err)
		return
	}
	app.Run()
}
