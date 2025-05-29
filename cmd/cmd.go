package cmd

import (
	_ "github.com/sikalabs/slr/cmd/allocate_100mb"
	_ "github.com/sikalabs/slr/cmd/client_ip_web_server"
	_ "github.com/sikalabs/slr/cmd/cpu_load_generator"
	_ "github.com/sikalabs/slr/cmd/dela"
	_ "github.com/sikalabs/slr/cmd/example"
	_ "github.com/sikalabs/slr/cmd/get_gps_info_from_jpg"
	_ "github.com/sikalabs/slr/cmd/get_helm_chart_version_from_repo"
	_ "github.com/sikalabs/slr/cmd/get_jwt_from_oidc"
	_ "github.com/sikalabs/slr/cmd/gitlab_create_branch"
	_ "github.com/sikalabs/slr/cmd/gitlab_create_merge_request"
	_ "github.com/sikalabs/slr/cmd/gitlab_update_file"
	_ "github.com/sikalabs/slr/cmd/gitlab_update_files"
	_ "github.com/sikalabs/slr/cmd/gojekyll"
	_ "github.com/sikalabs/slr/cmd/kubeconfig_from_vault"
	_ "github.com/sikalabs/slr/cmd/memory_usage"
	_ "github.com/sikalabs/slr/cmd/parse_jwt"
	_ "github.com/sikalabs/slr/cmd/redis_set_large_data"
	"github.com/sikalabs/slr/cmd/root"
	_ "github.com/sikalabs/slr/cmd/test_clisso_cli"
	_ "github.com/sikalabs/slr/cmd/validate_jwt"
	_ "github.com/sikalabs/slr/cmd/version"
	"github.com/spf13/cobra"
)

func Execute() {
	cobra.CheckErr(root.Cmd.Execute())
}
