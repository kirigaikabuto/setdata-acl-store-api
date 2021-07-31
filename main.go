package main

import (
	"fmt"
	"github.com/djumanoff/amqp"
	"github.com/joho/godotenv"
	setdata_acl "github.com/kirigaikabuto/setdata-acl"
	setdata_common "github.com/kirigaikabuto/setdata-common"
	"github.com/urfave/cli"
	"os"
	"strconv"
)

var (
	configPath           = ".env"
	version              = "0.0.0"
	amqpHost             = ""
	amqpPort             = 0
	amqpLevel            = ""
	postgresUser         = ""
	postgresPassword     = ""
	postgresDatabaseName = ""
	postgresHost         = ""
	postgresPort         = 5432
	postgresParams       = ""
	flags                = []cli.Flag{
		&cli.StringFlag{
			Name:        "config, c",
			Usage:       "path to .env config file",
			Destination: &configPath,
		},
	}
)

func parseEnvFile() {
	// Parse config file (.env) if path to it specified and populate env vars
	if configPath != "" {
		godotenv.Overload(configPath)
	}
	amqpHost = os.Getenv("RABBIT_HOST")
	amqpPortStr := os.Getenv("RABBIT_PORT")
	amqpPort, _ = strconv.Atoi(amqpPortStr)
	if amqpPort == 0 {
		amqpPort = 5432
	}
	if amqpHost == "" {
		amqpHost = "localhost"
	}
	postgresUser = os.Getenv("POSTGRES_USER")
	postgresPassword = os.Getenv("POSTGRES_PASSWORD")
	postgresDatabaseName = os.Getenv("POSTGRES_DATABASE")
	postgresParams = os.Getenv("POSTGRES_PARAMS")
	portStr := os.Getenv("POSTGRES_PORT")
	postgresPort, _ = strconv.Atoi(portStr)
	postgresHost = os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		postgresHost = "localhost"
	}
	if postgresPort == 0 {
		postgresPort = 5432
	}
	if postgresUser == "" {
		postgresUser = "setdatauser"
	}
	if postgresPassword == "" {
		postgresPassword = "123456789"
	}
	if postgresDatabaseName == "" {
		postgresDatabaseName = "setdata"
	}
	if postgresParams == "" {
		postgresParams = "sslmode=disable"
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Set Data Acl Store Api"
	app.Description = ""
	app.Usage = "set data run"
	app.UsageText = "set data run"
	app.Version = version
	app.Flags = flags
	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func run(c *cli.Context) error {
	parseEnvFile()
	rabbitConfig := amqp.Config{
		Host:     "localhost",
		Port:     5672,
		LogLevel: 5,
	}
	serverConfig := amqp.ServerConfig{
		ResponseX: "response",
		RequestX:  "request",
	}
	sess := amqp.NewSession(rabbitConfig)
	err := sess.Connect()
	if err != nil {
		return err
	}
	srv, err := sess.Server(serverConfig)
	if err != nil {
		return err
	}
	config := setdata_acl.PostgresConfig{
		Host:     postgresHost,
		Port:     postgresPort,
		User:     postgresUser,
		Password: postgresPassword,
		Database: postgresDatabaseName,
		Params:   postgresParams,
	}
	//role
	postgreRoleStore, err := setdata_acl.NewPostgresRoleStore(config)
	if err != nil {
		return err
	}
	roleService := setdata_acl.NewRoleService(postgreRoleStore)
	roleAmqpEndpoints := setdata_acl.NewRoleAmqpEndpoints(setdata_common.NewCommandHandler(roleService))
	srv.Endpoint("role.create", roleAmqpEndpoints.MakeCreateRoleAmqpEndpoint())
	srv.Endpoint("role.get", roleAmqpEndpoints.MakeGetRoleAmqpEndpoint())
	srv.Endpoint("role.list", roleAmqpEndpoints.MakeListRoleAmqpEndpoint())
	srv.Endpoint("role.delete", roleAmqpEndpoints.MakeDeleteRoleAmqpEndpoint())

	//permissions
	postgrePermissionStore, err := setdata_acl.NewPostgresPermissionStore(config)
	if err != nil {
		return err
	}
	permissionService := setdata_acl.NewPermissionService(postgrePermissionStore)
	permissionAmqpEndpoints := setdata_acl.NewPermissionAmqpEndpoints(setdata_common.NewCommandHandler(permissionService))
	srv.Endpoint("permission.create", permissionAmqpEndpoints.MakeCreatePermissionAmqpEndpoint())
	srv.Endpoint("permission.get", permissionAmqpEndpoints.MakeGetPermissionAmqpEndpoint())
	srv.Endpoint("permission.list", permissionAmqpEndpoints.MakeListPermissionAmqpEndpoint())
	srv.Endpoint("permission.delete", permissionAmqpEndpoints.MakeDeletePermissionAmqpEndpoint())

	//role permissions
	postgreRolePermissionStore, err := setdata_acl.NewPostgresRolePermissionStore(config)
	if err != nil {
		return err
	}
	rolePermissionService := setdata_acl.NewRolePermissionService(
		postgreRolePermissionStore,
	)
	rolePermissionAmqpEndpoints := setdata_acl.NewRolePermissionAmqpEndpoints(setdata_common.NewCommandHandler(rolePermissionService))
	srv.Endpoint("role_permission.create", rolePermissionAmqpEndpoints.MakeCreateRolePermissionAmqpEndpoint())
	srv.Endpoint("role_permission.list", rolePermissionAmqpEndpoints.MakeListRolePermissionAmqpEndpoint())
	srv.Endpoint("role_permission.get", rolePermissionAmqpEndpoints.MakeGetRolePermissionAmqpEndpoint())
	srv.Endpoint("role_permission.delete", rolePermissionAmqpEndpoints.MakeDeleteRolePermissionAmqpEndpoint())

	//user roles
	postgreUserRoleStore, err := setdata_acl.NewPostgresUserRoleStore(config)
	if err != nil {
		return err
	}
	userRoleService := setdata_acl.NewUserRoleService(postgreUserRoleStore)
	userRoleAmqpEndpoints := setdata_acl.NewUserRoleAmqpEndpoints(setdata_common.NewCommandHandler(userRoleService))
	srv.Endpoint("user_role.create", userRoleAmqpEndpoints.MakeCreateUserRoleAmqpEndpoint())
	srv.Endpoint("user_role.get", userRoleAmqpEndpoints.MakeGetUserRoleAmqpEndpoint())
	srv.Endpoint("user_role.delete", userRoleAmqpEndpoints.MakeDeleteUserRoleAmqpEndpoint())
	srv.Endpoint("user_role.list", userRoleAmqpEndpoints.MakeListUserRoleAmqpEndpoint())
	err = srv.Start()
	if err != nil {
		return err
	}
	return nil
}
