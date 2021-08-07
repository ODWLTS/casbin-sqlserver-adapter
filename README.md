# casbin-sqlserver-adapter

[![Go Report Card](https://goreportcard.com/badge/github.com/odwlts/casbin-sqlserver-adapter)](https://goreportcard.com/report/github.com/odwlts/casbin-sqlserver-adapter)
[![Coverage Status](https://coveralls.io/repos/github/odwlts/casbin-sqlserver-adapter/badge.svg?branch=master)](https://coveralls.io/github/odwlts/casbin-sqlserver-adapter?branch=master)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/odwlts/casbin-sqlserver-adapter)](https://pkg.go.dev/github.com/odwlts/casbin-sqlserver-adapter)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
---

casbin-sqlserver-adapter is a [Sqlx](https://github.com/jmoiron/sqlx) Adapter targeting [github.com/denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb) for [Casbin V2](https://github.com/casbin/casbin). 

With this library, Casbin can load policy lines from Sqlx supported databases or save policy lines.


## Tested Databases
- SQL Server(v2008R2-SP3): [github.com/denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb)

## Installation

	go get github.com/odwlts/casbin-sqlserver-adapter

## Getting Help

- [Casbin](https://github.com/casbin/casbin)


## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
