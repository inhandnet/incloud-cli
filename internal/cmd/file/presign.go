package file

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdPresign(f *factory.Factory) *cobra.Command {
	var filename string

	cmd := &cobra.Command{
		Use:   "presign",
		Short: "Generate a pre-signed URL for file upload",
		Long:  "Generate a pre-signed URL that can be used to upload files to the platform.",
		Example: `  # Generate a pre-signed URL
  incloud file presign

  # Generate a pre-signed URL for a specific filename
  incloud file presign --filename artifact.tar.gz`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if filename != "" {
				q.Set("filename", filename)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/files/presign", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&filename, "filename", "", "Filename for the pre-signed URL")

	return cmd
}
