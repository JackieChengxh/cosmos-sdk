package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

const (
	flagEpoch        = "epoch"
	flagLatestHeight = "latest-height"
)

// GetCmdQueryClientStates defines the command to query all the light clients
// that this chain mantains.
func GetCmdQueryClientStates() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "states",
		Short:   "Query all available light clients",
		Long:    "Query all available light clients",
		Example: fmt.Sprintf("%s query %s %s states", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryClientStatesRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.ClientStates(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "client states")

	return cmd
}

// GetCmdQueryClientState defines the command to query the state of a client with
// a given id as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryClientState() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "state [client-id]",
		Short:   "Query a client state",
		Long:    "Query stored client state",
		Example: fmt.Sprintf("%s query %s %s state [client-id]", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]
			prove, _ := cmd.Flags().GetBool(flags.FlagProve)

			clientStateRes, err := utils.QueryClientState(clientCtx, clientID, prove)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(clientStateRes)
		},
	}

	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryConsensusStates defines the command to query all the consensus states from a given
// client state.
func GetCmdQueryConsensusStates() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "consensus-states [client-id]",
		Short:   "Query all the consensus states of a client.",
		Long:    "Query all the consensus states from a given client state.",
		Example: fmt.Sprintf("%s query %s %s consensus-states [client-id]", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryConsensusStatesRequest{
				ClientId:   clientID,
				Pagination: pageReq,
			}

			res, err := queryClient.ConsensusStates(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "consensus states")

	return cmd
}

// GetCmdQueryConsensusState defines the command to query the consensus state of
// the chain as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryConsensusState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consensus-state [client-id] [height]",
		Short: "Query the consensus state of a client at a given height",
		Long: `Query the consensus state for a particular light client at a given height.
If the '--latest' flag is included, the query returns the latest consensus state, overriding the height argument.`,
		Example: fmt.Sprintf("%s query %s %s  consensus-state [client-id] [height]", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]

			queryLatestHeight, _ := cmd.Flags().GetBool(flagLatestHeight)

			var height uint64

			if !queryLatestHeight {
				if len(args) != 2 {
					return errors.New("must include a second 'height' argument when '--latest-height' flag is not provided")
				}

				height, err = strconv.ParseUint(args[1], 10, 64)
				if err != nil {
					return fmt.Errorf("expected integer height, got: %s", args[1])
				}
			}

			epoch, err := cmd.Flags().GetInt(flagEpoch)
			if err != nil {
				return fmt.Errorf("expected integer epoch: %w", err)
			}

			prove, _ := cmd.Flags().GetBool(flags.FlagProve)

			clientHeight := exported.NewHeight(uint64(epoch), height)
			csRes, err := utils.QueryConsensusState(clientCtx, clientID, clientHeight, prove, queryLatestHeight)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(csRes)
		},
	}

	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	cmd.Flags().Int(flagEpoch, 0, "epoch for the consensus state")
	cmd.Flags().Bool(flagLatestHeight, false, "return latest stored consensus state")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryHeader defines the command to query the latest header on the chain
func GetCmdQueryHeader() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "header",
		Short:   "Query the latest header of the running chain",
		Long:    "Query the latest Tendermint header of the running chain",
		Example: fmt.Sprintf("%s query %s %s  header", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			header, height, err := utils.QueryTendermintHeader(clientCtx)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutputLegacy(header)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdNodeConsensusState defines the command to query the latest consensus state of a node
// The result is feed to client creation
func GetCmdNodeConsensusState() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node-state",
		Short:   "Query a node consensus state",
		Long:    "Query a node consensus state. This result is feed to the client creation transaction.",
		Example: fmt.Sprintf("%s query %s %s node-state", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			state, height, err := utils.QueryNodeConsensusState(clientCtx)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(state)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
