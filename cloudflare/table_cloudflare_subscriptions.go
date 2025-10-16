package cloudflare

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/accounts"
	"github.com/cloudflare/cloudflare-go/v4/shared"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableCloudflareSubscriptions(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "cloudflare_subscriptions",
		Description: "",
		List: &plugin.ListConfig{
			Hydrate:       listSubscriptions,
			ParentHydrate: listAccount, 
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "account_id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			// Top columns
			{Name: "id", Type: proto.ColumnType_STRING, Transform: transform.FromField("ID"), Description: "Subscription identifier tag."},
			{Name: "currency", Type: proto.ColumnType_STRING, Description: "The monetary unit in which pricing information is displayed."},
			{Name: "current_period_end", Type: proto.ColumnType_TIMESTAMP, Description: "The end of the current period and also when the next billing is due."},
			{Name: "current_period_start", Type: proto.ColumnType_TIMESTAMP, Description: "When the current billing period started. May match initial_period_start if this is the first period."},
			{Name: "frequency", Type: proto.ColumnType_STRING, Description: "How often the subscription is renewed automatically."},
			{Name: "price", Type: proto.ColumnType_DOUBLE, Description: "The price of the subscription that will be billed, in US dollars."},
			{Name: "rate_plan", Type: proto.ColumnType_JSON, Description: "The rate plan applied to the subscription."},
			{Name: "state", Type: proto.ColumnType_STRING, Description: "The state that the subscription is in."},
		
			// Query columns for filtering
			{Name: "account_id", Type: proto.ColumnType_STRING, Transform: transform.FromField("AccountID"), Description: "The account ID to filter subscriptions."},
		
		}),
	}
}

type SubscriptionInfo struct {
	AccountID string
	shared.Subscription
}

//// LIST FUNCTION

// listSubscriptions retrieves all subscriptions across all accounts.
func listSubscriptions(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	accountDetails := h.Item.(accounts.Account)

	inputAccountId := d.EqualsQualString("account_id")

	// Only list subscription for accounts stated in the input query
	if inputAccountId != "" && inputAccountId != accountDetails.ID {
		return nil, nil
	}

	conn, err := connectV4(ctx, d)
	if err != nil {
		logger.Error("cloudflare_subscription.listSubscriptions", "connect error", err)
		return nil, err
	}

	input := accounts.SubscriptionGetParams{
		AccountID: cloudflare.F(accountDetails.ID),
	}

    page, err := conn.Accounts.Subscriptions.Get(ctx, input)
    if err != nil {
        logger.Error("cloudflare_subscription.listSubscriptions", "api call error", err)
        return nil, err
    }

    if page == nil {
        return nil, nil
    }

    subs := page.Result 
    for _, curr := range subs {
        sub := SubscriptionInfo{
            AccountID:    accountDetails.ID,
            Subscription: curr,
        }
        d.StreamListItem(ctx, sub)

        if d.RowsRemaining(ctx) == 0 {
            return nil, nil
        }
    }

    return nil, nil
}
