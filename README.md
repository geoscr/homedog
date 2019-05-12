# Homedog
Monitors Craigslist and Kijiji for posts matching a seach filter set up on the source website. Matching posts are emailed in realtime, but duplicate posts/renews are ignored using a fuzzy text match that catches common repost variants. 

To take an example, we're trying to find a 2 bedroom apartment for July 1, not basement or condo, 5km radius from given coordinates except for Hochelaga and Outremont, which are specific areas of town we want to avoid.

That's easy to do by starting with Craigslist and Kijiji's built-in filters from their website (provided as RSS links, a service offered by both providers), and supplemented with a few 'removal' keywords that flag a post as definitely not interesting, for example:

    "may 1*",
    "june 1*",
    "hochelaga",
    "outremont",
    "condo",
    "basement",
    "sous[- ]sol"         # bilingual support
    ( "2e", "triplex" )
    ( "2nd", "triplex" )  # combine matches - don't want to live in middle floor of a 3 storey building
    
In this way, obvious negative matches are filtered out, and potential positives are emailed almost in real-time (subject to provider API rate-limiting), allowing an effort (of keeping up with the email stream) to be rewarded by being (often) the first person to call about the post.

## Installation guide

### Prerequisites

- POSIX-compliant host
- Docker

First create a .env file then a config file on the host at `../config/config.json` (pointed to by the symlink `config` in the root of the repository).

### Example `.env`

    POSTGRES_PASSWORD=

    HOMEDOG_AWS_REGION=
    HOMEDOG_AWS_KEY=
    HOMEDOG_AWS_SECRET=

    HOMEDOG_SMTP_HOST=smtp.example.com
    HOMEDOG_SMTP_PORT=587

    HOMEDOG_SENDER=homedog@example.com
    HOMEDOG_CONFIG=/app/config/config.json

### Example `config.json`

    {
        "Subscribers": [
            {
                "Email": "user@tp1.org", 
                "Properties": {
                    "HasPic": 1,
                    "Min_price": 1500,
                    "Max_price": 2000,
                    "Min_bedrooms": 1,
                    "Max_bedrooms": 2,
                    "Postal": "H3K 1G6",
                    "Coordinates": "45.4741399,-73.5813671",
                    "Search_distance": 2,
                    "Furnished": 0,
                    "Exclusions": [ "june 1", "basement", "sous-sol" ]
                }
            }
            /* repeat subscriber block as needed */
        ]
    }

## Running

    docker-compose up
