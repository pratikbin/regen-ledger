package server

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/query"

	api "github.com/regen-network/regen-ledger/api/regen/data/v1"
	"github.com/regen-network/regen-ledger/x/data"
)

func TestQuery_ResolversByHash(t *testing.T) {
	t.Parallel()
	s := setupBase(t)

	id1 := []byte{0}
	ch1 := &data.ContentHash{Graph: &data.ContentHash_Graph{
		Hash:                      bytes.Repeat([]byte{0}, 32),
		DigestAlgorithm:           data.DigestAlgorithm_DIGEST_ALGORITHM_BLAKE2B_256,
		CanonicalizationAlgorithm: data.GraphCanonicalizationAlgorithm_GRAPH_CANONICALIZATION_ALGORITHM_URDNA2015,
	}}
	iri1, err := ch1.ToIRI()
	require.NoError(t, err)

	id2 := []byte{1}
	ch2 := &data.ContentHash{Graph: &data.ContentHash_Graph{
		Hash:                      bytes.Repeat([]byte{1}, 32),
		DigestAlgorithm:           data.DigestAlgorithm_DIGEST_ALGORITHM_BLAKE2B_256,
		CanonicalizationAlgorithm: data.GraphCanonicalizationAlgorithm_GRAPH_CANONICALIZATION_ALGORITHM_URDNA2015,
	}}
	iri2, err := ch2.ToIRI()
	require.NoError(t, err)

	// insert data ids
	err = s.server.stateStore.DataIDTable().Insert(s.ctx, &api.DataID{Id: id1, Iri: iri1})
	require.NoError(t, err)
	err = s.server.stateStore.DataIDTable().Insert(s.ctx, &api.DataID{Id: id2, Iri: iri2})
	require.NoError(t, err)

	// insert resolvers
	rid1, err := s.server.stateStore.ResolverTable().InsertReturningID(s.ctx, &api.Resolver{
		Url:     testURL,
		Manager: s.addrs[0],
	})
	require.NoError(t, err)
	rid2, err := s.server.stateStore.ResolverTable().InsertReturningID(s.ctx, &api.Resolver{
		Url:     testURL,
		Manager: s.addrs[1],
	})
	require.NoError(t, err)

	// insert registration records
	err = s.server.stateStore.DataResolverTable().Insert(s.ctx, &api.DataResolver{
		Id:         id1,
		ResolverId: rid1,
	})
	require.NoError(t, err)
	err = s.server.stateStore.DataResolverTable().Insert(s.ctx, &api.DataResolver{
		Id:         id1,
		ResolverId: rid2,
	})
	require.NoError(t, err)

	// query resolvers with valid content hash
	res, err := s.server.ResolversByHash(s.ctx, &data.QueryResolversByHashRequest{
		ContentHash: ch1,
		Pagination:  &query.PageRequest{Limit: 1, CountTotal: true},
	})
	require.NoError(t, err)

	// check pagination
	require.Len(t, res.Resolvers, 1)
	require.Equal(t, uint64(2), res.Pagination.Total)

	// check resolver properties
	require.Equal(t, rid1, res.Resolvers[0].Id)
	require.Equal(t, s.addrs[0].String(), res.Resolvers[0].Manager)
	require.Equal(t, testURL, res.Resolvers[0].Url)

	// query resolvers with content hash that has not been registered
	res, err = s.server.ResolversByHash(s.ctx, &data.QueryResolversByHashRequest{
		ContentHash: ch2,
	})
	require.NoError(t, err)
	require.Empty(t, res.Resolvers)

	// query resolvers with empty content hash
	_, err = s.server.ResolversByHash(s.ctx, &data.QueryResolversByHashRequest{})
	require.EqualError(t, err, "content hash cannot be empty: invalid request")

	// query resolvers with invalid content hash
	_, err = s.server.ResolversByHash(s.ctx, &data.QueryResolversByHashRequest{
		ContentHash: &data.ContentHash{},
	})
	require.EqualError(t, err, "invalid data.ContentHash: invalid type")

	// query resolvers with content hash that has not been anchored
	_, err = s.server.ResolversByHash(s.ctx, &data.QueryResolversByHashRequest{
		ContentHash: &data.ContentHash{Graph: &data.ContentHash_Graph{
			Hash:                      bytes.Repeat([]byte{2}, 32),
			DigestAlgorithm:           data.DigestAlgorithm_DIGEST_ALGORITHM_BLAKE2B_256,
			CanonicalizationAlgorithm: data.GraphCanonicalizationAlgorithm_GRAPH_CANONICALIZATION_ALGORITHM_URDNA2015,
		}},
	})
	require.EqualError(t, err, "data record with content hash: not found")
}
