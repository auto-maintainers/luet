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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mudler/luet/pkg/compiler"
	"github.com/mudler/luet/pkg/compiler/backend"
	. "github.com/mudler/luet/pkg/config"
	. "github.com/mudler/luet/pkg/logger"
	pkg "github.com/mudler/luet/pkg/package"
	tree "github.com/mudler/luet/pkg/tree"

	_gentoo "github.com/Sabayon/pkgs-checker/pkg/gentoo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:   "build <package name> <package name> <package name> ...",
	Short: "build a package or a tree",
	Long:  `build packages or trees from luet tree definitions. Packages are in [category]/[name]-[version] form`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("clean", cmd.Flags().Lookup("clean"))
		viper.BindPFlag("tree", cmd.Flags().Lookup("tree"))
		viper.BindPFlag("destination", cmd.Flags().Lookup("destination"))
		viper.BindPFlag("backend", cmd.Flags().Lookup("backend"))
		viper.BindPFlag("privileged", cmd.Flags().Lookup("privileged"))
		viper.BindPFlag("database", cmd.Flags().Lookup("database"))
		viper.BindPFlag("revdeps", cmd.Flags().Lookup("revdeps"))
		viper.BindPFlag("all", cmd.Flags().Lookup("all"))
		viper.BindPFlag("compression", cmd.Flags().Lookup("compression"))
		viper.BindPFlag("nodeps", cmd.Flags().Lookup("nodeps"))
		viper.BindPFlag("onlydeps", cmd.Flags().Lookup("onlydeps"))

		viper.BindPFlag("image-repository", cmd.Flags().Lookup("image-repository"))
		viper.BindPFlag("push", cmd.Flags().Lookup("push"))
		viper.BindPFlag("pull", cmd.Flags().Lookup("pull"))
		viper.BindPFlag("keep-images", cmd.Flags().Lookup("keep-images"))

		LuetCfg.Viper.BindPFlag("keep-exported-images", cmd.Flags().Lookup("keep-exported-images"))

		LuetCfg.Viper.BindPFlag("solver.type", cmd.Flags().Lookup("solver-type"))
		LuetCfg.Viper.BindPFlag("solver.discount", cmd.Flags().Lookup("solver-discount"))
		LuetCfg.Viper.BindPFlag("solver.rate", cmd.Flags().Lookup("solver-rate"))
		LuetCfg.Viper.BindPFlag("solver.max_attempts", cmd.Flags().Lookup("solver-attempts"))
	},
	Run: func(cmd *cobra.Command, args []string) {

		clean := viper.GetBool("clean")
		src := viper.GetString("tree")
		dst := viper.GetString("destination")
		concurrency := LuetCfg.GetGeneral().Concurrency
		backendType := viper.GetString("backend")
		privileged := viper.GetBool("privileged")
		revdeps := viper.GetBool("revdeps")
		all := viper.GetBool("all")
		databaseType := viper.GetString("database")
		compressionType := viper.GetString("compression")
		imageRepository := viper.GetString("image-repository")
		push := viper.GetBool("push")
		pull := viper.GetBool("pull")
		keepImages := viper.GetBool("keep-images")
		nodeps := viper.GetBool("nodeps")
		onlydeps := viper.GetBool("onlydeps")
		keepExportedImages := viper.GetBool("keep-exported-images")

		compilerSpecs := compiler.NewLuetCompilationspecs()
		var compilerBackend compiler.CompilerBackend
		var db pkg.PackageDatabase
		switch backendType {
		case "img":
			compilerBackend = backend.NewSimpleImgBackend()
		case "docker":
			compilerBackend = backend.NewSimpleDockerBackend()
		}

		switch databaseType {
		case "memory":
			db = pkg.NewInMemoryDatabase(false)

		case "boltdb":
			tmpdir, err := ioutil.TempDir("", "package")
			if err != nil {
				Fatal(err)
			}
			db = pkg.NewBoltDatabase(tmpdir)

		}
		defer db.Clean()

		generalRecipe := tree.NewCompilerRecipe(db)

		Info("Loading", src)
		Info("Building in", dst)

		err := generalRecipe.Load(src)
		if err != nil {
			Fatal("Error: " + err.Error())
		}

		stype := LuetCfg.Viper.GetString("solver.type")
		discount := LuetCfg.Viper.GetFloat64("solver.discount")
		rate := LuetCfg.Viper.GetFloat64("solver.rate")
		attempts := LuetCfg.Viper.GetInt("solver.max_attempts")

		LuetCfg.GetSolverOptions().Type = stype
		LuetCfg.GetSolverOptions().LearnRate = float32(rate)
		LuetCfg.GetSolverOptions().Discount = float32(discount)
		LuetCfg.GetSolverOptions().MaxAttempts = attempts

		Debug("Solver", LuetCfg.GetSolverOptions().CompactString())

		opts := compiler.NewDefaultCompilerOptions()
		opts.SolverOptions = *LuetCfg.GetSolverOptions()
		opts.ImageRepository = imageRepository
		opts.Clean = clean
		opts.PullFirst = pull
		opts.KeepImg = keepImages
		opts.Push = push
		opts.OnlyDeps = onlydeps
		opts.NoDeps = nodeps
		opts.KeepImageExport = keepExportedImages

		luetCompiler := compiler.NewLuetCompiler(compilerBackend, generalRecipe.GetDatabase(), opts)
		luetCompiler.SetConcurrency(concurrency)
		luetCompiler.SetCompressionType(compiler.CompressionImplementation(compressionType))
		if !all {
			for _, a := range args {
				gp, err := _gentoo.ParsePackageStr(a)
				if err != nil {
					Fatal("Invalid package string ", a, ": ", err.Error())
				}

				if gp.Version == "" {
					gp.Version = "0"
					gp.Condition = _gentoo.PkgCondGreaterEqual
				}

				pack := &pkg.DefaultPackage{
					Name: gp.Name,
					Version: fmt.Sprintf("%s%s%s",
						pkg.PkgSelectorConditionFromInt(gp.Condition.Int()).String(),
						gp.Version,
						gp.VersionSuffix,
					),
					Category: gp.Category,
					Uri:      make([]string, 0),
				}
				spec, err := luetCompiler.FromPackage(pack)
				if err != nil {
					Fatal("Error: " + err.Error())
				}

				spec.SetOutputPath(dst)
				compilerSpecs.Add(spec)
			}
		} else {
			w := generalRecipe.GetDatabase().World()

			for _, p := range w {
				spec, err := luetCompiler.FromPackage(p)
				if err != nil {
					Fatal("Error: " + err.Error())
				}
				Info(":package: Selecting ", p.GetName(), p.GetVersion())
				spec.SetOutputPath(dst)
				compilerSpecs.Add(spec)
			}
		}

		var artifact []compiler.Artifact
		var errs []error
		if revdeps {
			artifact, errs = luetCompiler.CompileWithReverseDeps(privileged, compilerSpecs)

		} else {
			artifact, errs = luetCompiler.CompileParallel(privileged, compilerSpecs)

		}
		if len(errs) != 0 {
			for _, e := range errs {
				Error("Error: " + e.Error())
			}
			Fatal("Bailing out")
		}
		for _, a := range artifact {
			Info("Artifact generated:", a.GetPath())
		}
	},
}

func init() {
	path, err := os.Getwd()
	if err != nil {
		Fatal(err)
	}
	buildCmd.Flags().Bool("clean", true, "Build all packages without considering the packages present in the build directory")
	buildCmd.Flags().String("tree", path, "Source luet tree")
	buildCmd.Flags().String("backend", "docker", "backend used (docker,img)")
	buildCmd.Flags().Bool("privileged", false, "Privileged (Keep permissions)")
	buildCmd.Flags().String("database", "memory", "database used for solving (memory,boltdb)")
	buildCmd.Flags().Bool("revdeps", false, "Build with revdeps")
	buildCmd.Flags().Bool("all", false, "Build all packages in the tree")
	buildCmd.Flags().String("destination", path, "Destination folder")
	buildCmd.Flags().String("compression", "none", "Compression alg: none, gzip")
	buildCmd.Flags().String("image-repository", "luet/cache", "Default base image string for generated image")
	buildCmd.Flags().Bool("push", false, "Push images to a hub")
	buildCmd.Flags().Bool("pull", false, "Pull images from a hub")
	buildCmd.Flags().Bool("keep-images", true, "Keep built docker images in the host")
	buildCmd.Flags().Bool("nodeps", false, "Build only the target packages, skipping deps (it works only if you already built the deps locally, or by using --pull) ")
	buildCmd.Flags().Bool("onlydeps", false, "Build only package dependencies")
	buildCmd.Flags().Bool("keep-exported-images", false, "Keep exported images used during building")

	buildCmd.Flags().String("solver-type", "", "Solver strategy")
	buildCmd.Flags().Float32("solver-rate", 0.7, "Solver learning rate")
	buildCmd.Flags().Float32("solver-discount", 1.0, "Solver discount rate")
	buildCmd.Flags().Int("solver-attempts", 9000, "Solver maximum attempts")

	RootCmd.AddCommand(buildCmd)
}
