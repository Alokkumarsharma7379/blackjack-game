package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Card struct {
	Rank string
	Suit string
}

func (c Card) Value() int {
	switch c.Rank {
	case "J", "Q", "K":
		return 10
	case "A":
		return 11
	default:
		v, _ := strconv.Atoi(c.Rank)
		return v
	}
}

func (c Card) String() string {
	return c.Rank + c.Suit
}

type Deck struct {
	cards []Card
}

func NewDeck() *Deck {
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	suits := []string{"♠", "♥", "♦", "♣"}
	d := &Deck{}
	for i := 0; i < 4; i++ { // 4 decks (208 cards) for a bit more realism
		for _, s := range suits {
			for _, r := range ranks {
				d.cards = append(d.cards, Card{Rank: r, Suit: s})
			}
		}
	}
	d.Shuffle()
	return d
}

func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

func (d *Deck) Draw() Card {
	if len(d.cards) == 0 {
		*d = *NewDeck()
	}
	c := d.cards[len(d.cards)-1]
	d.cards = d.cards[:len(d.cards)-1]
	return c
}

type Hand struct {
	Cards []Card
}

func (h *Hand) Add(c Card) {
	h.Cards = append(h.Cards, c)
}

func (h Hand) Value() int {
	total := 0
	aces := 0
	for _, c := range h.Cards {
		total += c.Value()
		if c.Rank == "A" {
			aces++
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

func (h Hand) IsBlackjack() bool {
	return len(h.Cards) == 2 && h.Value() == 21
}

func (h Hand) IsBust() bool {
	return h.Value() > 21
}

func (h Hand) String(hideFirst bool) string {
	var parts []string
	for i, c := range h.Cards {
		if hideFirst && i == 0 {
			parts = append(parts, "??")
		} else {
			parts = append(parts, c.String())
		}
	}
	return strings.Join(parts, " ")
}

type Game struct {
	deck    *Deck
	reader  *bufio.Reader
	balance int
}

func NewGame() *Game {
	return &Game{
		deck:    NewDeck(),
		reader:  bufio.NewReader(os.Stdin),
		balance: 1000,
	}
}

func (g *Game) prompt(msg string) string {
	fmt.Print(msg)
	text, _ := g.reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func (g *Game) getBet() int {
	for {
		input := g.prompt(fmt.Sprintf("Balance: $%d. Enter your bet: ", g.balance))
		bet, err := strconv.Atoi(input)
		if err != nil || bet <= 0 {
			fmt.Println("Please enter a valid positive number.")
			continue
		}
		if bet > g.balance {
			fmt.Println("You can't bet more than your balance.")
			continue
		}
		return bet
	}
}

func (g *Game) playRound() {
	bet := g.getBet()

	if len(g.deck.cards) < 20 {
		g.deck = NewDeck()
	}

	player := &Hand{}
	dealer := &Hand{}

	player.Add(g.deck.Draw())
	dealer.Add(g.deck.Draw())
	player.Add(g.deck.Draw())
	dealer.Add(g.deck.Draw())

	fmt.Println()
	fmt.Println(strings.Repeat("-", 40))
	g.showTable(player, dealer, true)

	// Check for blackjacks
	playerBJ := player.IsBlackjack()
	dealerBJ := dealer.IsBlackjack()

	if playerBJ || dealerBJ {
		g.showTable(player, dealer, false)
		switch {
		case playerBJ && dealerBJ:
			fmt.Println("Both have Blackjack! Push.")
		case playerBJ:
			winnings := int(float64(bet) * 1.5)
			fmt.Printf("Blackjack! You win $%d!\n", winnings)
			g.balance += winnings
		case dealerBJ:
			fmt.Println("Dealer has Blackjack! You lose your bet.")
			g.balance -= bet
		}
		return
	}

	// Player's turn
	for {
		if player.IsBust() {
			break
		}
		choice := strings.ToLower(g.prompt("Hit or Stand? (h/s): "))
		if choice == "h" || choice == "hit" {
			player.Add(g.deck.Draw())
			g.showTable(player, dealer, true)
		} else if choice == "s" || choice == "stand" {
			break
		} else {
			fmt.Println("Please enter 'h' or 's'.")
		}
	}

	if player.IsBust() {
		g.showTable(player, dealer, false)
		fmt.Println("Bust! You lose.")
		g.balance -= bet
		return
	}

	// Dealer's turn (hits until 17+, stands on soft 17)
	fmt.Println("\nDealer's turn...")
	for dealer.Value() < 17 {
		dealer.Add(g.deck.Draw())
	}
	g.showTable(player, dealer, false)

	pVal, dVal := player.Value(), dealer.Value()
	switch {
	case dealer.IsBust():
		fmt.Printf("Dealer busts! You win $%d!\n", bet)
		g.balance += bet
	case pVal > dVal:
		fmt.Printf("You win $%d! (%d vs %d)\n", bet, pVal, dVal)
		g.balance += bet
	case pVal < dVal:
		fmt.Printf("You lose. (%d vs %d)\n", pVal, dVal)
		g.balance -= bet
	default:
		fmt.Println("Push! Bet returned.")
	}
}

func (g *Game) showTable(player, dealer *Hand, hideDealer bool) {
	fmt.Printf("Dealer: %s", dealer.String(hideDealer))
	if !hideDealer {
		fmt.Printf(" (%d)", dealer.Value())
	}
	fmt.Println()
	fmt.Printf("You:    %s (%d)\n", player.String(false), player.Value())
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("=== Welcome to CLI Blackjack ===")

	g := NewGame()

	for {
		if g.balance <= 0 {
			fmt.Println("\nYou're out of money. Game over!")
			break
		}

		g.playRound()
		fmt.Printf("\nBalance: $%d\n", g.balance)

		again := strings.ToLower(g.prompt("\nPlay another round? (y/n): "))
		if again != "y" && again != "yes" {
			break
		}
	}

	fmt.Printf("\nThanks for playing! Final balance: $%d\n", g.balance)
}
