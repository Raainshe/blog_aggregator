#!/bin/bash

# Script to populate database with dummy data
# Creates 10 users, each with 2 feeds

echo "Starting database population..."

# Array of dummy users
users=(
    "alice_dev"
    "bob_blogger" 
    "charlie_tech"
    "diana_writer"
    "eve_coder"
    "frank_reviewer"
    "grace_editor"
    "henry_author"
    "iris_creator"
    "jack_publisher"
)

# Array of feed names and URLs (2 per user)
feed_names=(
    "Tech News" "Programming Tips"
    "Web Development" "JavaScript Weekly"
    "Python Blog" "Data Science"
    "Creative Writing" "Book Reviews"
    "Code Tutorials" "Software Reviews"
    "Tech Reviews" "Gadget News"
    "Content Creation" "Writing Tips"
    "Tech Industry" "Startup Stories"
    "Design Blog" "UI/UX Tips"
    "Publishing News" "Editorial Insights"
)

feed_urls=(
    "https://techcrunch.com/feed" "https://blog.codinghorror.com/rss"
    "https://css-tricks.com/feed" "https://javascriptweekly.com/rss"
    "https://realpython.com/feed" "https://towardsdatascience.com/feed"
    "https://writerswrite.co.za/feed" "https://www.goodreads.com/blog/rss"
    "https://dev.to/feed" "https://softwareengineeringdaily.com/feed"
    "https://www.theverge.com/rss/index.xml" "https://www.engadget.com/rss.xml"
    "https://contentmarketinginstitute.com/feed" "https://thewritepractice.com/feed"
    "https://venturebeat.com/feed" "https://techstartups.com/feed"
    "https://www.smashingmagazine.com/feed" "https://uxplanet.org/feed"
    "https://www.publishersweekly.com/rss" "https://www.editorandpublisher.com/feed"
)

# Function to add feeds for a user
add_feeds_for_user() {
    local user=$1
    local start_idx=$2
    
    echo "Adding feeds for user: $user"
    
    # Login as the user first
    go run . login "$user"
    
    # Add 2 feeds for this user
    local feed1_name="${feed_names[$start_idx]}"
    local feed1_url="${feed_urls[$start_idx]}"
    local feed2_name="${feed_names[$((start_idx + 1))]}"
    local feed2_url="${feed_urls[$((start_idx + 1))]}"
    
    echo "Adding feed 1: $feed1_name"
    go run . addfeed "$feed1_name" "$feed1_url"
    
    echo "Adding feed 2: $feed2_name"
    go run . addfeed "$feed2_name" "$feed2_url"
    
    echo "Completed feeds for $user"
    echo "---"
}

# Main loop
for i in "${!users[@]}"; do
    user="${users[$i]}"
    feed_idx=$((i * 2))
    
    echo "Registering user: $user"
    go run . register "$user"
    
    # Add feeds for this user
    add_feeds_for_user "$user" "$feed_idx"
done

echo "Database population completed!"
echo "Summary:"
echo "- Registered ${#users[@]} users"
echo "- Added $(( ${#users[@]} * 2 )) feeds total"

# Show final state
echo ""
echo "Current users:"
go run . users
