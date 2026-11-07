package stars

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

const allQuery = `query($cursor: String) {
  viewer {
    starredRepositories(first: 100, after: $cursor, orderBy: {field: STARRED_AT, direction: DESC}) {
      nodes { url }
      pageInfo { hasNextPage endCursor }
    }
  }
}`

const listsQuery = `query($cursor: String) {
  viewer {
    lists(first: 100, after: $cursor) {
      nodes { id name }
      pageInfo { hasNextPage endCursor }
    }
  }
}`

const itemsQuery = `query($id: ID!, $cursor: String) {
  node(id: $id) {
    ... on UserList {
      items(first: 100, after: $cursor) {
        nodes { ... on Repository { url } }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}`

type Client struct {
	api *api.GraphQLClient
}

func NewClient() (*Client, error) {
	client, err := api.NewGraphQLClient(api.ClientOptions{Timeout: 30 * time.Second})
	if err != nil {
		return nil, err
	}
	return &Client{api: client}, nil
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type connection[T any] struct {
	Nodes    []*T     `json:"nodes"`
	PageInfo pageInfo `json:"pageInfo"`
}

type repository struct {
	URL string `json:"url"`
}

type userList struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// walk fetches every page, rejecting malformed cursors to avoid infinite loops.
func walk[T any](fetch func(*string) (connection[T], error), visit func(T) error) error {
	var cursor *string
	seen := make(map[string]bool)
	for {
		page, err := fetch(cursor)
		if err != nil {
			return err
		}
		for _, node := range page.Nodes {
			if node != nil {
				if err := visit(*node); err != nil {
					return err
				}
			}
		}
		if !page.PageInfo.HasNextPage {
			return nil
		}
		next := page.PageInfo.EndCursor
		if next == "" || seen[next] {
			return fmt.Errorf("invalid pagination cursor returned by GitHub")
		}
		seen[next] = true
		cursor = &next
	}
}

func printRepository(out io.Writer) func(repository) error {
	return func(repo repository) error {
		if repo.URL == "" {
			return nil
		}
		_, err := fmt.Fprintln(out, repo.URL)
		return err
	}
}

func (c *Client) All(ctx context.Context, out io.Writer) error {
	return walk(func(cursor *string) (connection[repository], error) {
		var result struct {
			Viewer struct {
				Repositories connection[repository] `json:"starredRepositories"`
			} `json:"viewer"`
		}
		err := c.api.DoWithContext(ctx, allQuery, map[string]any{"cursor": cursor}, &result)
		return result.Viewer.Repositories, err
	}, printRepository(out))
}

func (c *Client) eachList(ctx context.Context, visit func(userList) error) error {
	return walk(func(cursor *string) (connection[userList], error) {
		var result struct {
			Viewer struct {
				Lists connection[userList] `json:"lists"`
			} `json:"viewer"`
		}
		err := c.api.DoWithContext(ctx, listsQuery, map[string]any{"cursor": cursor}, &result)
		return result.Viewer.Lists, err
	}, visit)
}

func (c *Client) Lists(ctx context.Context, out io.Writer) error {
	return c.eachList(ctx, func(list userList) error {
		_, err := fmt.Fprintln(out, list.Name)
		return err
	})
}

func (c *Client) List(ctx context.Context, name string, out io.Writer) error {
	var id string
	err := c.eachList(ctx, func(list userList) error {
		if list.Name == name {
			if id != "" {
				return fmt.Errorf("multiple lists named %q", name)
			}
			id = list.ID
		}
		return nil
	})
	if err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("list not found: %q", name)
	}
	return walk(func(cursor *string) (connection[repository], error) {
		var result struct {
			Node *struct {
				Items connection[repository] `json:"items"`
			} `json:"node"`
		}
		err := c.api.DoWithContext(ctx, itemsQuery, map[string]any{"id": id, "cursor": cursor}, &result)
		if err != nil {
			return connection[repository]{}, err
		}
		if result.Node == nil {
			return connection[repository]{}, fmt.Errorf("list %q is no longer available", name)
		}
		return result.Node.Items, nil
	}, printRepository(out))
}
