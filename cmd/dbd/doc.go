/*
dbd is a simple database server.  It registers an endpoint that
can be used to retrieve data to/from it's database.

For example
    $ go build && ./dbd --http :55555 &
    $ curl "http://127.0.0.1:55555/list?type=cards&user=user1@test.com&q=*"
    [
        {
            "id": 8,
            "owner": "user2@test.com:algebra",
            "front": "x+x",
            "back": "2x"
        },
        {
            "id": 9,
            "owner": "user2@test.com:programming",
            "front": "favorite programming language",
            "back": "Go"
        },
        {
            "id": 10,
            "owner": "user2@test.com:programming",
            "front": "public interface",
            "back": "API"
        }
    ]

    $ curl "http://127.0.0.1:55555/list?type=decks&user=admin&q=*"
    [
        {
            "name": "user1@test.com.com:spanish"
        },
        {
            "name": "user2@test.com:algebra"
        },
        {
            "name": "user2@test.com:programming"
        }
    ]

    $ curl "http://127.0.0.1:55555/list?type=users&user=admin&q=*"
    [
        {
            "email": "user1@test.com.com",
            "name": "User1",
            "password": "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"
        },
        {
            "email": "user2@test.com",
            "name": "User2",
            "password": "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"
        }
    ]

    $  curl -X POST -H "Content-Type: application/json" -d '[{"owner": "user2@test.com:numbers", "front": "x+x", "back": "2x"}]

The database operates on Card, Deck, and User types.  Users own Decks which are
made up of Cards.

*/
package main
