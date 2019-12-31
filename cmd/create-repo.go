// Copyright © 2019 Ettore Di Giacinto <mudler@gentoo.org>
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
package cmd

import (
	"os"

	installer "github.com/mudler/luet/pkg/installer"

	. "github.com/mudler/luet/pkg/logger"
	pkg "github.com/mudler/luet/pkg/package"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createrepoCmd = &cobra.Command{
	Use:   "create-repo",
	Short: "Create a luet repository from a build",
	Long:  `Generate and renew repository metadata`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("packages", cmd.Flags().Lookup("packages"))
		viper.BindPFlag("tree", cmd.Flags().Lookup("tree"))
		viper.BindPFlag("output", cmd.Flags().Lookup("output"))
		viper.BindPFlag("name", cmd.Flags().Lookup("name"))
		viper.BindPFlag("descr", cmd.Flags().Lookup("descr"))
		viper.BindPFlag("urls", cmd.Flags().Lookup("urls"))
		viper.BindPFlag("type", cmd.Flags().Lookup("type"))
	},
	Run: func(cmd *cobra.Command, args []string) {

		tree := viper.GetString("tree")
		dst := viper.GetString("output")
		packages := viper.GetString("packages")
		name := viper.GetString("name")
		descr := viper.GetString("descr")
		urls := viper.GetStringSlice("urls")
		t := viper.GetString("type")

		repo, err := installer.GenerateRepository(name, descr, t, urls, 1, packages, tree, pkg.NewInMemoryDatabase(false))
		if err != nil {
			Fatal("Error: " + err.Error())
		}
		err = repo.Write(dst)
		if err != nil {
			Fatal("Error: " + err.Error())
		}
	},
}

func init() {
	path, err := os.Getwd()
	if err != nil {
		Fatal(err)
	}
	createrepoCmd.Flags().String("packages", path, "Packages folder (output from build)")
	createrepoCmd.Flags().String("tree", path, "Source luet tree")
	createrepoCmd.Flags().String("output", path, "Destination folder")
	createrepoCmd.Flags().String("name", "luet", "Repository name")
	createrepoCmd.Flags().String("descr", "luet", "Repository description")
	createrepoCmd.Flags().StringSlice("urls", []string{}, "Repository URLs")
	createrepoCmd.Flags().String("type", "disk", "Repository type (disk)")

	RootCmd.AddCommand(createrepoCmd)
}
