// Copyright © 2020 Ettore Di Giacinto <mudler@gentoo.org>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, see <http://www.gnu.org/licenses/>.

package cmd_database

import (
	"io/ioutil"
	"path/filepath"

	"github.com/mudler/luet/pkg/compiler"
	. "github.com/mudler/luet/pkg/logger"
	pkg "github.com/mudler/luet/pkg/package"

	. "github.com/mudler/luet/pkg/config"

	"github.com/spf13/cobra"
)

func NewDatabaseCreateCommand() *cobra.Command {
	var ans = &cobra.Command{
		Use:   "create <artifact_metadata1.yaml> <artifact_metadata1.yaml>",
		Short: "Insert a package in the system DB",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			LuetCfg.Viper.BindPFlag("system.database_path", cmd.Flags().Lookup("system-dbpath"))
			LuetCfg.Viper.BindPFlag("system.rootfs", cmd.Flags().Lookup("system-target"))

		},
		Run: func(cmd *cobra.Command, args []string) {

			var systemDB pkg.PackageDatabase

			for _, a := range args {
				dat, err := ioutil.ReadFile(a)
				if err != nil {
					Fatal("Failed reading ", a, ": ", err.Error())
				}
				art, err := compiler.NewPackageArtifactFromYaml(dat)
				if err != nil {
					Fatal("Failed reading yaml ", a, ": ", err.Error())
				}

				if LuetCfg.GetSystem().DatabaseEngine == "boltdb" {
					systemDB = pkg.NewBoltDatabase(
						filepath.Join(LuetCfg.GetSystem().GetSystemRepoDatabaseDirPath(), "luet.db"))
				} else {
					systemDB = pkg.NewInMemoryDatabase(true)
				}

				files := art.GetFiles()

				if _, err := systemDB.CreatePackage(art.GetCompileSpec().GetPackage()); err != nil {
					Fatal("Failed to create ", a, ": ", err.Error())
				}
				if err := systemDB.SetPackageFiles(&pkg.PackageFile{PackageFingerprint: art.GetCompileSpec().GetPackage().GetFingerPrint(), Files: files}); err != nil {
					Fatal("Failed setting package files for ", a, ": ", err.Error())
				}

				Info(art.GetCompileSpec().GetPackage().HumanReadableString(), " created")
			}

		},
	}

	return ans
}
