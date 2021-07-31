package main

import (
	"fmt"
	"github.com/djumanoff/amqp"
	setdata_acl "github.com/kirigaikabuto/setdata-acl"
	setdata_common "github.com/kirigaikabuto/setdata-common"
)

func main() {
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
		panic(err)
		return
	}
	srv, err := sess.Server(serverConfig)
	if err != nil {
		panic(err)
		return
	}
	config := setdata_acl.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "setdatauser",
		Password: "123456789",
		Database: "setdata",
		Params:   "sslmode=disable",
	}
	//role
	postgreRoleStore, err := setdata_acl.NewPostgresRoleStore(config)
	if err != nil {
		panic(err)
		return
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
		panic(err)
		return
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
		panic(err)
		return
	}
	rolePermissionService := setdata_acl.NewRolePermissionService(
		postgreRolePermissionStore,
		postgreRoleStore,
		postgrePermissionStore,
	)
	rolePermissionAmqpEndpoints := setdata_acl.NewRolePermissionAmqpEndpoints(setdata_common.NewCommandHandler(rolePermissionService))
	srv.Endpoint("role_permission.create", rolePermissionAmqpEndpoints.MakeCreateRolePermissionAmqpEndpoint())
	srv.Endpoint("role_permission.list", rolePermissionAmqpEndpoints.MakeListRolePermissionAmqpEndpoint())
	srv.Endpoint("role_permission.get", rolePermissionAmqpEndpoints.MakeGetRolePermissionAmqpEndpoint())
	srv.Endpoint("role_permission.delete", rolePermissionAmqpEndpoints.MakeDeleteRolePermissionAmqpEndpoint())

	fmt.Println(postgreRolePermissionStore)
	err = srv.Start()
	if err != nil {
		panic(err)
		return
	}

}
