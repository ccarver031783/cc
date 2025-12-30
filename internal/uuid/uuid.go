package uuid

import (
	"fmt"

	"github.com/google/uuid"
	ufcli "github.com/urfave/cli/v2"
)

func NewUUIDCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "uuid",
		Usage: "Generate a UUID v4",
		Flags: []ufcli.Flag{
			&ufcli.IntFlag{
				Name:  "count",
				Usage: "Generate a specific number of UUIDs",
				Value: 1,
			},
		},
		Action: func(c *ufcli.Context) error {
			count := c.Int("count")
			for i := 0; i < count; i++ {
				id := uuid.New()
				fmt.Println(id.String())
			}
			return nil
		},
	}
}
