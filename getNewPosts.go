package main

import (
	"fmt"
	embed "github.com/Clinet/discordgo-embed"
	"github.com/enescakir/emoji"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"regexp"
	"strconv"
	"strings"
)

var maxId = make(map[string]int)

func makeRequest(qpath string, data []byte) (res *abci.ResponseQuery, err error) {
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	remote := "gno.land:36657"
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(qpath, data, opts2)
	if err != nil {
		return nil, err
	}
	if qres.Response.Error != nil {
		fmt.Printf("Log: %s\n",
			qres.Response.Log)
		return nil, qres.Response.Error
	}
	return &qres.Response, nil
}

func getBoardsPosts(board string) (string, error) {
	qpath := "vm/qrender"
	data := []byte(fmt.Sprintf("%s\n%s", "gno.land/r/boards", board))
	res, err := makeRequest(qpath, data)

	if err != nil {
		fmt.Println("Error: ", res.Log)
		return "", err
	}
	return string(res.Data), nil
}

type Post struct {
	Title       string
	Author      string
	Description string
	Id          int
}

func GetPostInfos(post string, id int) Post {
	regAuthor := regexp.MustCompile(`\\- \[(@[a-z]+)\]`)
	regTitle := regexp.MustCompile(`## \[([^\[\]]+)\]`)
	regDescription := regexp.MustCompile(`(?s)\)\n\n.*\n\\`)
	matchTitle := regTitle.FindStringSubmatch(post)
	matchAuthor := regAuthor.FindStringSubmatch(post)
	matchDescription := regDescription.FindStringSubmatch(post)[0][3:]
	matchDescription = matchDescription[:len(matchDescription)-2]

	fmt.Println(post)

	p := Post{
		Title:       matchTitle[1],
		Author:      matchAuthor[1],
		Description: matchDescription,
		Id:          id,
	}
	fmt.Println(p)
	return p
}

func GetPostID(s string) (int, error) {
	re := regexp.MustCompile("\\bpostid=([0-9]+)")
	match := re.FindStringSubmatch(s)
	if len(match) == 0 {
		return 0, nil
	}
	return strconv.Atoi(match[1])
}

func parseNewPosts(BoardPosts string, board string) []*embed.Embed {
	var post []Post
	newMaxId := maxId[board]
	a := strings.Split(BoardPosts, "----------------------------------------")
	for _, c := range a {
		nb, _ := GetPostID(c)
		if nb > maxId[board] {
			post = append(post, GetPostInfos(c, nb))
			if nb > newMaxId {
				newMaxId = nb
			}
		}
	}
	maxId[board] = newMaxId
	return EmbedNewPosts(post, board)
}

func EmbedNewPosts(posts []Post, board string) []*embed.Embed {
	embeds := make([]*embed.Embed, 0)
	for _, post := range posts {
		embeds = append(embeds, embed.NewEmbed().
			SetTitle(fmt.Sprintf("New post on: %s %v ", board, emoji.OpenMailboxWithRaisedFlag)).
			SetDescription(fmt.Sprintf("**%s**\n%s\n\nhttps://gno.land/r/boards:%s/%d", post.Title, post.Description, board, post.Id)).
			SetAuthor(post.Author).
			SetColor(0x6e0e08))
	}
	fmt.Printf("THERE IS %d NEW POSTS\n", len(embeds))
	return embeds
}

func getNewPosts(board string) ([]*embed.Embed, error) {
	// this return the posts from the watched board
	BoardPosts, err := getBoardsPosts(board)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile("\\bpostid=([0-9]+)")
	var newIdString = re.FindAllStringSubmatch(BoardPosts, -1)
	// var newId []int

	for _, i := range newIdString {
		j, err := strconv.Atoi(i[1])
		if err != nil {
			panic(err)
		}
		if j > maxId[board] {
			return parseNewPosts(BoardPosts, board), nil
		}
	}
	return nil, nil
}
