package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

var Parties = []Game{}
var Cards = []Card{Card{0, "Boule de feu", "ATK", 5}, Card{1, "Barriere", "DEF", 4}, Card{2, "Bandage", "HEAL", 3}, Card{3, "Patate de forrain", "ATK", 6}, Card{4, "Bouclier", "DEF", 2}, Card{5, "Herbe", "HEAL", 3}, Card{6, "Epée du Lord", "ATK", 8}, Card{7, "Armure du Lord", "DEF", 6}, Card{8, "Prière du Lord", "HEAL", 7}, Card{9, "Pointe de glace", "ATK", 3}, Card{10, "Mur de glace", "DEF", 4}, Card{11, "Pluie de guérison", "HEAL", 2}}
var Heroes = []Hero{}

type Response struct {
	GameId int
	Player Hero
}

type Game struct {
	Id            int
	Players       []Hero
	Winner, Loser Hero
	State         string
	//Other thing...
}

type Hero struct {
	Id     int
	Name   string
	Pv     int
	Hand   []Card
	Choice Card
}

type Card struct {
	Id             int
	Name, CardType string
	Effect         int
}

func (h *Hero) choose(c int) {
	h.Choice = h.Hand[c]
	fmt.Println(h.Name, " Choosed ", h.Choice.Name)
}

func fight(h1, h2 Hero) (Hero, Hero) {

	c1 := h1.Choice
	c2 := h2.Choice

	switch c1.CardType {
	case "DEF":
		h1.Pv += c1.Effect
	case "HEAL":
		h1.Pv += c1.Effect
	case "ATK":
		h2.Pv -= c1.Effect
	}

	switch c2.CardType {
	case "DEF":
		h2.Pv += c2.Effect
	case "HEAL":
		h2.Pv += c2.Effect
	case "ATK":
		h1.Pv -= c2.Effect
	}

	h1.Choice = Card{}
	h2.Choice = Card{}

	return h1, h2
}

func randomCards() []Card {
	heroCards := []Card{}
	rand.Seed(time.Now().UnixNano())

	i := 0

	for i < 3 {

		randomIndex := rand.Intn(len(Cards))
		pick := Cards[randomIndex]
		fmt.Println(randomIndex)
		heroCards = append(heroCards, pick)
		i++
	}

	return heroCards
}

func (p *Game) giveCardsToHeroes() {
	for i, _ := range p.Players {
		p.Players[i].Hand = randomCards()
	}
}

func addNewPartyIntoParties(channel chan Game) {
	newGame := Game{len(Parties), []Hero{}, Hero{}, Hero{}, "not started"}

	party := append(Parties, newGame)
	Parties = party

	channel <- newGame
}

func createParty(chanl chan int) {
	for {
		newGame := Game{len(Parties), []Hero{}, Hero{}, Hero{}, "not started"}

		party := append(Parties, newGame)
		Parties = party

		msg := <-chanl
		fmt.Println("La partie a été créée", msg)
	}
	// channel := make(chan Game)
	// go addNewPartyIntoParties(channel)

	// party := <-channel

	// if len(party.players) > 0 {
	// 	if !party.isPlayerOne() || !party.isPlayerTwo() || !party.isPlayerOneAndTwo() {
	// 		fmt.Println("Need Hero to start a game")
	// 		//Party created but there is one  or two players missing -> add player 1 or 2
	// 	} else {
	// 		//Party starting
	// 		fmt.Println("The game is starting...")

	// 		party.giveCardsToHeroes()
	// 	}
	// } else {
	// 	// Party need players -> add player to game
	// 	fmt.Println("Add hero to game")
	// }
}

func (p *Game) isPlayerOne() bool {
	fmt.Println("playersss", p.Players)
	return p.Players[0].Name != ""
}

func (p *Game) isPlayerTwo() bool {
	return p.Players[1].Name != ""
}

func (p *Game) isPlayerOneAndTwo() bool {
	return p.isPlayerOne() && p.isPlayerTwo()
}

func createPartyHandler(channel chan int) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// get response headers
		header := w.Header()

		// set content type header
		header.Set("Content-Type", "application/json")

		// respond with a JSON string
		b, err := json.Marshal(Parties[len(Parties)-1])

		if err != nil {
			w.WriteHeader(500)
			fmt.Println("error", err)
			fmt.Fprint(w, string(b))
		} else {
			w.WriteHeader(200)
			fmt.Println(string(b))
			fmt.Fprint(w, string(b))
			channel <- len(Parties)
		}

	}
}

func addPlayerToPartyHandler(channel chan Response) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// get response headers
		header := w.Header()

		// set content type header
		header.Set("Content-Type", "application/json")

		// respond with a JSON string
		var h Hero

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&h)
		if err != nil {
			w.WriteHeader(500)
			fmt.Println("error", err)
			fmt.Fprint(w, err)
		} else {
			w.WriteHeader(200)
			fmt.Println("Hero:", h)
			fmt.Fprintf(w, "Hero: %+v", h)
			i, _ := strconv.Atoi(ps.ByName("id"))
			channel <- Response{i, Hero{len(Heroes), h.Name, 25, []Card{}, Card{}}}
		}
	}
}

func addPlayerToParty(channel chan Response) {
	for {
		msg := <-channel

		Heroes = append(Heroes, msg.Player)
		for i, v := range Parties {
			if v.Id == msg.GameId {
				fmt.Println("founded", v)
				v.Players = append(v.Players, msg.Player)
				fmt.Println("edited", v)
				Parties[i] = v
			}
		}

		fmt.Println("Player added to Party", msg)
	}
}

func getParties(channel chan int) {
	for {
		msg := <-channel
		fmt.Println("Number of parties", msg)
	}
}

func getPartiesHandler(channel chan int) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// get response headers
		header := w.Header()

		// set content type header
		header.Set("Content-Type", "application/json")

		// respond with a JSON string
		b, err := json.Marshal(Parties)

		if err != nil {
			w.WriteHeader(500)
			fmt.Println("error", err)
			fmt.Fprint(w, string(b))
		} else {
			w.WriteHeader(200)
			fmt.Println(string(b))
			fmt.Fprint(w, string(b))
			channel <- len(Parties)
		}

	}
}

func startParty(channel chan int) {
	for {
		msg := <-channel
		for i, v := range Parties {
			if v.Id == msg {
				v.State = "started"
				v.giveCardsToHeroes()
				fmt.Print(v)
				Parties[i] = v
			}
		}
		fmt.Println("Games", msg, "is starting")
	}
}

func startPartyHandler(channel chan int) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// get response headers
		header := w.Header()

		// set content type header
		header.Set("Content-Type", "application/json")

		// respond with a JSON string

		w.WriteHeader(200)
		fmt.Println("Game is starting")
		fmt.Fprint(w, "Game is starting")

		i, _ := strconv.Atoi(ps.ByName("id"))
		channel <- i

	}
}

func heroMakeChoice(channel chan []int) {
	for {
		msg := <-channel
		for _, v := range Parties {
			if v.Id == msg[2] {
				for j, l := range v.Players {
					if l.Id == msg[0] {
						v.Players[j].choose(msg[1])
					}
				}
			}
		}
		fmt.Println("Hero make his choice", msg)
	}
}

func heroMakeChoiceHandler(channel chan []int) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// get response headers
		header := w.Header()

		// set content type header
		header.Set("Content-Type", "application/json")

		// respond with a JSON string

		w.WriteHeader(200)
		fmt.Println("Hero", ps.ByName("playerId"), "make choice number", ps.ByName("choiceNb"))
		fmt.Fprint(w, "Hero ", ps.ByName("playerId"), " make choice number ", ps.ByName("choiceNb"))

		i, _ := strconv.Atoi(ps.ByName("playerId"))
		j, _ := strconv.Atoi(ps.ByName("choiceNb"))
		k, _ := strconv.Atoi(ps.ByName("gameId"))
		channel <- []int{i, j, k}

	}
}

func heroesFight(channel chan int) {
	for {
		msg := <-channel
		for i, v := range Parties {
			if v.Id == msg {
				v.Players[0], v.Players[1] = fight(v.Players[0], v.Players[1])
				v.giveCardsToHeroes()
				Parties[i] = v

			}
		}

		fmt.Println("Heroes fight", msg)
	}
}

func heroesFightHandler(channel chan int) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// get response headers
		header := w.Header()

		// set content type header
		header.Set("Content-Type", "application/json")

		// respond with a JSON string

		w.WriteHeader(200)
		fmt.Println("Heroes fights")
		fmt.Fprint(w, "Heroes fights")

		i, _ := strconv.Atoi(ps.ByName("id"))
		channel <- i

	}
}

func main() {

	// var scanner = bufio.NewScanner(os.Stdin)

	channel, channel2, channel3, channel4, channel5, channel6 := make(chan int), make(chan Response), make(chan int), make(chan int), make(chan []int), make(chan int)

	go createParty(channel)
	go addPlayerToParty(channel2)
	go getParties(channel3)
	go startParty(channel4)
	go heroMakeChoice(channel5)
	go heroesFight(channel6)

	router := httprouter.New()
	router.GET("/party/create", createPartyHandler(channel))
	router.POST("/games/:id/player/add", addPlayerToPartyHandler(channel2))
	router.GET("/parties", getPartiesHandler(channel3))
	router.POST("/parties/:id/start", startPartyHandler(channel4))
	router.POST("/heroes/:playerId/parties/:gameId/choices/:choiceNb", heroMakeChoiceHandler(channel5))
	router.POST("/parties/:id/fight", heroesFightHandler(channel6))

	log.Fatal(http.ListenAndServe(":8080", router))

	// fmt.Println(Cards, h1, h2)

	// fmt.Println("INIT")
	// fmt.Println(h1)
	// fmt.Println(h2)

	// var Choice1 int

	// fmt.Print(h1.Name, " Make your Choice : ")
	// scanner.Scan()
	// Choice1, err := strconv.Atoi(scanner.Text())
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// h1.choose(Choice1)
	// fmt.Println(h1.Choice)
	// var Choice2 int

	// fmt.Print(h2.Name, " Make your Choice : ")
	// scanner.Scan()
	// Choice2, err = strconv.Atoi(scanner.Text())
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// h2.choose(Choice2)

	// h1, h2 = fight(h1, h2)

	// fmt.Println("ROUND 0")

	// fmt.Println(h1)
	// fmt.Println(h2)
}
