// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"fmt"
	"strings"

	"github.com/posener/complete"
)

type ACLTokenUpdateCommand struct {
	Meta

	roleNames []string
	roleIDs   []string
}

func (c *ACLTokenUpdateCommand) Help() string {
	helpText := `
Usage: nomad acl token update <token_accessor_id>

  Update is used to update an existing ACL token. Requires a management token.

General Options:

  ` + generalOptionsUsage(usageOptsDefault|usageOptsNoNamespace) + `

Update Options:

  -name=""
    Sets the human readable name for the ACL token.

  -type="client"
    Sets the type of token. Must be one of "client" or "management".

  -policy=""
    Specifies a policy to associate with the token. Can be specified multiple
    times, but only with client type tokens. If any policies are specified, they
    completely replace the policies on the existing token.

  -role-id=""
    ID of a role to use for this token. Can be specified multiple times, but
    only with client type tokens. If any roles are specified, they completely
    replace the roles on the existing token.

  -role-name=""
    Name of a role to use for this token. Can be specified multiple times, but
    only with client type tokens. If any roles are specified, they completely
    replace the roles on the existing token.
`

	return strings.TrimSpace(helpText)
}

func (c *ACLTokenUpdateCommand) AutocompleteFlags() complete.Flags {
	return mergeAutocompleteFlags(c.Meta.AutocompleteFlags(FlagSetClient),
		complete.Flags{
			"name":      complete.PredictAnything,
			"type":      complete.PredictAnything,
			"policy":    complete.PredictAnything,
			"role-id":   complete.PredictAnything,
			"role-name": complete.PredictAnything,
		})
}

func (c *ACLTokenUpdateCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *ACLTokenUpdateCommand) Synopsis() string {
	return "Update an existing ACL token"
}

func (*ACLTokenUpdateCommand) Name() string { return "acl token update" }

func (c *ACLTokenUpdateCommand) Run(args []string) int {
	var name, tokenType string
	var policies []string
	flags := c.Meta.FlagSet(c.Name(), FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.StringVar(&name, "name", "", "")
	flags.StringVar(&tokenType, "type", "", "")
	flags.Var((funcVar)(func(s string) error {
		policies = append(policies, s)
		return nil
	}), "policy", "")
	flags.Var((funcVar)(func(s string) error {
		c.roleNames = append(c.roleNames, s)
		return nil
	}), "role-name", "")
	flags.Var((funcVar)(func(s string) error {
		c.roleIDs = append(c.roleIDs, s)
		return nil
	}), "role-id", "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	// Check that we got exactly one argument
	args = flags.Args()
	if l := len(args); l != 1 {
		c.Ui.Error("This command takes one argument: <token_accessor_id>")
		c.Ui.Error(commandErrorText(c))
		return 1
	}

	tokenAccessorID := args[0]

	// Get the HTTP client
	client, err := c.Meta.Client()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing client: %s", err))
		return 1
	}

	// Get the specified token
	token, _, err := client.ACLTokens().Info(tokenAccessorID, nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error fetching token: %s", err))
		return 1
	}

	// Create the updated token
	if name != "" {
		token.Name = name
	}

	if tokenType != "" {
		token.Type = tokenType
	}

	if len(policies) != 0 {
		token.Policies = policies
	}

	if len(c.roleNames) != 0 || len(c.roleIDs) != 0 {
		token.Roles = generateACLTokenRoleLinks(c.roleNames, c.roleIDs)
	}

	// Update the token
	updatedToken, _, err := client.ACLTokens().Update(token, nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error updating token: %s", err))
		return 1
	}

	// Format the output
	outputACLToken(c.Ui, updatedToken)
	return 0
}
