package docread

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

const testSample = `
## Uh oh

# Webhook Resource

Webhooks are a low-effort way to post messages to channels in Discord. They do not require a bot user or authentication to use.

### Webhook Object

Used to represent a webhook.

###### Webhook Structure

| Field              | Type                                                             | Description                                                                                                   |
| ------------------ | ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| id                 | snowflake                                                        | the id of the webhook                                                                                         |
| type               | integer                                                          | the [type](#DOCS_RESOURCES_WEBHOOK/webhook-object-webhook-types) of the webhook                               |
| guild_id?          | ?snowflake                                                       | the guild id this webhook is for, if any                                                                      |
| channel_id         | ?snowflake                                                       | the channel id this webhook is for, if any                                                                    |
| user?              | [user](#DOCS_RESOURCES_USER/user-object) object                  | the user this webhook was created by (not returned when getting a webhook with its token)                     |
| name               | ?string                                                          | the default name of the webhook                                                                               |
| avatar             | ?string                                                          | the default user avatar [hash](#DOCS_REFERENCE/image-formatting) of the webhook                               |
| token?             | string                                                           | the secure token of the webhook (returned for Incoming Webhooks)                                              |
| application_id     | ?snowflake                                                       | the bot/OAuth2 application that created this webhook                                                          |
| source_guild? *    | partial [guild](#DOCS_RESOURCES_GUILD/guild-object) object       | the guild of the channel that this webhook is following (returned for Channel Follower Webhooks)              |
| source_channel? *  | partial [channel](#DOCS_RESOURCES_CHANNEL/channel-object) object | the channel that this webhook is following (returned for Channel Follower Webhooks)                           |
| url?               | string                                                           | the url used for executing the webhook (returned by the [webhooks](#DOCS_TOPICS_OAUTH2/webhooks) OAuth2 flow) |

\* These fields will be absent if the webhook creator has since lost access to the guild where the followed channel resides

###### Webhook Types

| Value | Name             | Description                                                                                                    |
| ----- | ---------------- | -------------------------------------------------------------------------------------------------------------- |
| 1     | Incoming         | Incoming Webhooks can post messages to channels with a generated token                                         |
| 2     | Channel Follower | Channel Follower Webhooks are internal webhooks used with Channel Following to post new messages into channels |
| 3     | Application      | Application webhooks are webhooks used with Interactions                                                       |

###### Example Incoming Webhook
`

func TestScrapeBytes(t *testing.T) {
	tables, err := ScrapeBytes([]byte(testSample))
	assert.NoError(t, err)
	assert.Equal(t, []ObjectTable{
		{
			Sections: []string{"Webhook Resource", "Webhook Object", "Webhook Structure"},
			Table: []ObjectTableRow{
				{
					Field:       "id",
					Type:        "snowflake",
					Description: "the id of the webhook",
				},
				{
					Field:       "type",
					Type:        "integer",
					Description: "the [type](#DOCS_RESOURCES_WEBHOOK/webhook-object-webhook-types) of the webhook",
				},
				{
					Field:       "guild_id?",
					Type:        "?snowflake",
					Description: "the guild id this webhook is for, if any",
				},
				{
					Field:       "channel_id",
					Type:        "?snowflake",
					Description: "the channel id this webhook is for, if any",
				},
				{
					Field:       "user?",
					Type:        "[user](#DOCS_RESOURCES_USER/user-object) object",
					Description: "the user this webhook was created by (not returned when getting a webhook with its token)",
				},
				{
					Field:       "name",
					Type:        "?string",
					Description: "the default name of the webhook",
				},
				{
					Field:       "avatar",
					Type:        "?string",
					Description: "the default user avatar [hash](#DOCS_REFERENCE/image-formatting) of the webhook",
				},
				{
					Field:       "token?",
					Type:        "string",
					Description: "the secure token of the webhook (returned for Incoming Webhooks)",
				},
				{
					Field:       "application_id",
					Type:        "?snowflake",
					Description: "the bot/OAuth2 application that created this webhook",
				},
				{
					Field:       "source_guild?",
					Type:        "partial [guild](#DOCS_RESOURCES_GUILD/guild-object) object",
					Description: "the guild of the channel that this webhook is following (returned for Channel Follower Webhooks)",
				},
				{
					Field:       "source_channel?",
					Type:        "partial [channel](#DOCS_RESOURCES_CHANNEL/channel-object) object",
					Description: "the channel that this webhook is following (returned for Channel Follower Webhooks)",
				},
				{
					Field:       "url?",
					Type:        "string",
					Description: "the url used for executing the webhook (returned by the [webhooks](#DOCS_TOPICS_OAUTH2/webhooks) OAuth2 flow)",
				},
			},
			Source: Source{Position: 236},
		},
	}, tables)
}

func TestRelativeLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want float64
	}{
		{"", "", 1},
		{"", "a", 0},
		{"a", "", 0},
		{"a", "a", 1},
		{"aaaaaa gui ld aaa", "guild", 0.2941176470588235},
		{"guild", "aaaaaa gui ld aaa", 0.2941176470588235},
		{"asdasdasd", "zxczxczxc", 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s-%s", tt.a, tt.b), func(t *testing.T) {
			got := relativeLevenshtein(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
