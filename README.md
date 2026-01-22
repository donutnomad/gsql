# GSQL - Type-Safe SQL Query Builder for GORM

A type-safe, fluent SQL query builder library built on top of GORM with support for complex queries, CASE expressions, CTE, batch operations, and 100+ MySQL functions.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://go.dev/)
[![GORM](https://img.shields.io/badge/GORM-v1.31.0-green)](https://gorm.io/)

## Features

- ✅ Type-safe query building with Go generics
- ✅ Fluent, chainable API
- ✅ CASE WHEN expressions builder
- ✅ CTE (Common Table Expressions) support
- ✅ BatchIn optimizer for large IN queries
- ✅ 100+ MySQL functions wrapped
- ✅ Subqueries and JOINs
- ✅ JSON field operations

## Installation

```bash
go get github.com/donutnomad/gsql
```