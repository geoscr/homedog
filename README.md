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
    "sous[- ]sol"
    ( "2e", "triplex" )   # combine matches - don't want to live in middle floor of a 3 storey building
    
In this way, obvious negative matches are filtered out, and potential positives are emailed almost in real-time (subject to provider API rate-limiting), allowing an effort (of keeping up with the email stream) to be rewarded by being (often) the first person to call about the post.
