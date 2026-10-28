package migrations

import "embed"

// Миграции разложены по сервисам: у каждого своя схема (своя БД), FK только внутри неё.

//go:embed iam/*.sql
var IAM embed.FS

//go:embed catalog/*.sql
var Catalog embed.FS

//go:embed appointment/*.sql
var Appointment embed.FS

//go:embed payment/*.sql
var Payment embed.FS

//go:embed notification/*.sql
var Notification embed.FS
