// Copyright 2021 by ODW LTS. All Rights Reserved.
// Copyright 2020 by Blank-Xu. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file file has been modified by ODW LTS to improve SQL Server support.

package casbinsqlserveradapter

const (
	sqlCreateTable = `
CREATE TABLE %[1]s.%[2]s(
    p_type NVARCHAR(32)  DEFAULT '' NOT NULL,
    v0     NVARCHAR(255) DEFAULT '' NOT NULL,
    v1     NVARCHAR(255) DEFAULT '' NOT NULL,
    v2     NVARCHAR(255) DEFAULT '' NOT NULL,
    v3     NVARCHAR(255) DEFAULT '' NOT NULL,
    v4     NVARCHAR(255) DEFAULT '' NOT NULL,
    v5     NVARCHAR(255) DEFAULT '' NOT NULL
);
CREATE INDEX idx_%[2]s ON %[1]s.%[2]s (p_type, v0, v1);`

	sqlTruncateTable = "TRUNCATE TABLE %s.%s"
	sqlIsTableExist  = "SELECT 1 FROM %s.%s"
	sqlSelectAll     = "SELECT p_type,v0,v1,v2,v3,v4,v5 FROM %s.%s"
	sqlSelectWhere   = "SELECT p_type,v0,v1,v2,v3,v4,v5 FROM %s.%s WHERE "
	sqlInsertRow     = "INSERT INTO %s.%s (p_type,v0,v1,v2,v3,v4,v5) VALUES (?,?,?,?,?,?,?)"
	sqlUpdateRow     = "UPDATE %s.%s SET p_type=?,v0=?,v1=?,v2=?,v3=?,v4=?,v5=? WHERE p_type=? AND v0=? AND v1=? AND v2=? AND v3=? AND v4=? AND v5=?"
	sqlDeleteRow     = "DELETE FROM %s.%s WHERE p_type=? AND v0=? AND v1=? AND v2=? AND v3=? AND v4=? AND v5=?"
	sqlDeleteAll     = "DELETE FROM %s.%s"
	sqlDeleteByArgs  = "DELETE FROM %s.%s WHERE p_type=?"
)
