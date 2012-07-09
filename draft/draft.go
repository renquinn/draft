package main

import (
	"fmt"
	"net/http"
	"html/template"

	"strconv"
	"time"

	"appengine"
	"appengine/urlfetch"
	"appengine/datastore"

	"encoding/xml"
	"encoding/json"
	"io"
)

// ===============================
// Globals
// ===============================

// Server Management
// The list of templates to use
var TEMPLATES = template.Must(template.ParseFiles(
				"index.html",
				"head.html",
				"navbar.html",
				"footer.html",
				"lobby.html",
				"keepers.html",
				"admin.html",
				"test.html",
				"news.html",
				"research.html"))

// Default draft info (for now)
var NUMROUNDS int
var NUMTEAMS int
var CLOCK time.Time
var PAUSE bool

// Pick Management
//teamsCH := make(chan string, 12) // The channel by which to send picks // FUTURE: Use the App Engine Channel API
var PICKS[13][16]Player // The main draft data, each pick is stored here PICKS[NUMTEAMS+1][NUMROUNDS+1]
var ALLPICKS []Player
var PLAYERS Players // All players
var TEAMS []Team

// Team info
var ADMINS = []string {"dixie","el_gor"}
var teamPassword = map[string] string {
	"dixie": "sky67",
	"b_ez_on": "bra67",
	"up_n_at": "geo58",
	"i_am_ba": "cha58",
	"rob_do": "rob103",
	"bhers": "mat76",
	"el_gor": "ren76",
	"nativ": "tre49",
	"p_town": "jor103",
	"hit_sq": "kyl94",
	"impac": "con67",
	"ukrai": "tyl49",
	"test": "test123",
}
var teamNumber = map[string] int {
	"dixie": 12,
	"b_ez_on": 2,
	"up_n_at": 3,
	"i_am_ba": 4,
	"rob_do": 11,
	"bhers": 7,
	"el_gor": 6,
	"nativ": 1,
	"p_town": 9,
	"hit_sq": 10,
	"impac": 5,
	"ukrai": 8,
}
var teamName = map[string] string {
	"dixie": "Skyler",
	"b_ez_on": "Brady",
	"up_n_at": "Geoff",
	"i_am_ba": "Chad",
	"rob_do": "Rob",
	"bhers": "Matt",
	"el_gor": "Ren",
	"nativ": "Trevor",
	"p_town": "Jordan",
	"hit_sq": "Kyle",
	"impac": "Conner",
	"ukrai": "Tyler",
}
var teamTeam = map[string] string {
	"dixie": "Dixie",
	"b_ez_on": "B EZ ON MY SNAX",
	"up_n_at": "Up N Atoms",
	"i_am_ba": "I Am Batman",
	"rob_do": "Rob Dogo",
	"bhers": "BHers",
	"el_gor": "El Gordo",
	"nativ": "Native Americans",
	"p_town": "P-Town",
	"hit_sq": "Hit Squad",
	"impac": "Impact",
	"ukrai": "Ukraine",
}

// ===============================
// Structs
// ===============================
type Page struct {
	User User
	AllPicks []Player
	Rosters Players
	League []Team
	Pause string
	Players Players
}

type User struct {
	Name string
	Username string
	Team string
	Number int
	Picks [15]Player
}

func (u *User) Key(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "User", u.Username, 0, nil)
}

type Team struct {
	User string
	Name string
	Number int
	TabID string
	QB []Player
	RB []Player
	WR []Player
	TE []Player
	K []Player
	DEF []Player
}

type Players struct {
	QB map[string]Player
	RB map[string]Player
	WR map[string]Player
	TE map[string]Player
	K map[string]Player
	DEF map[string]Player
	ALL map[string]Player
}

/*
 * Template Structs
 */

type Head struct {
	Title string
	Pause bool
}

/*
 * XML Structs
 */

type Player struct {
	PlayerID string `xml:"playerId,attr"`
	Name string `datastore:"name" xml:"Name,attr"`
	Position string `xml:"Position,attr"`
	Team string `xml:"Team,attr"`
}

type FFN struct {
	XMLName xml.Name `xml:"FantasyFootballNerd"`
	Results []Player `xml:"Results>Player"`
}

/*
 * JSON Structs
 */

type News struct {
	Headlines []Headline
}

type Headline struct {
	Headline string
	Description string
	Links Links
	Categories []Category
}

type Links struct {
	Web Web
}

type Web struct {
	Href string
	Athletes Athletes
}

type Research struct {
	Headlines []Headline
}

type Category struct {
	Athlete Athlete
}

type Athlete struct {
	Id int
	Description string
	Links Links
}

type Athletes struct {
	Href string
}

// ===============================
// Helpers 
// ===============================

// Returns a User struct from the datastore
func getUser(c appengine.Context, username string) (*User, error) {
	u := &User{Username: username}
	err := datastore.Get(c,u.Key(c), u)
	if err == datastore.ErrNoSuchEntity {
		_, err = datastore.Put(c, u.Key(c), u)
	}
	return u, err
}

// Returns all the Users in the draft
func getUsers(c appengine.Context) ([]User, error) {
	k := make([]*datastore.Key,0)
	u := new(User)
	for i:=1;i<=12;i++ {
		u.Username = rlookup(teamNumber,i)
		k = append(k,u.Key(c))
	}
	users := make([]User, len(k))
	err := datastore.GetMulti(c,k,users)
	return users, err
}

// Returns a timer formatted in scoreboard style
func getTime() string {
	if PAUSE {
		return "00:00:00"
	}
	timer := int(time.Now().Sub(CLOCK).Seconds())
	minutes := timer/60
	seconds := timer%60
	hours := minutes/60
	minutes = minutes%60
	hzero := ""
	mzero := ""
	szero := ""
	if hours < 10 {
		hzero = "0"
	}
	if seconds < 10 {
		szero = "0"
	}
	if minutes < 10 {
		mzero = "0"
	}
	return hzero + strconv.Itoa(hours) + ":" + mzero + strconv.Itoa(minutes) + ":" + szero + strconv.Itoa(seconds)
}

// Allows reverse lookup for maps
func rlookup(m map[string] int, n int) string {
	for k,v := range m {
		if v == n {
			return k
		}
	}
	return ""
}

// Retrieves the rosters from fantasyfootballnerd.com and stores them in the datastore
func UpdateRosters(players []Player, r *http.Request) (error) {
	qb := make(map[string]Player)
	rb := make(map[string]Player)
	wr := make(map[string]Player)
	te := make(map[string]Player)
	k := make(map[string]Player)
	def := make(map[string]Player)
	all := make(map[string]Player)

	c := appengine.NewContext(r)

	for _,p := range players {

		key := datastore.NewKey(c, "player", p.PlayerID,0,nil)
		_, err := datastore.Put(c, key, &p)
		if err != nil {
			return err
		}

		switch p.Position {
		case "QB":
			//qb = append(qb, p)
			qb[p.PlayerID] = p
		case "RB":
			//rb = append(rb, p)
			rb[p.PlayerID] = p
		case "WR":
			//wr = append(wr, p)
			wr[p.PlayerID] = p
		case "TE":
			//te = append(te, p)
			te[p.PlayerID] = p
		case "K":
			//k = append(k, p)
			k[p.PlayerID] = p
		case "DEF":
			//def = append(def, p)
			def[p.PlayerID] = p
		}
		all[p.PlayerID] = p
	}
	PLAYERS = Players{QB:qb,RB:rb,WR:wr,TE:te,K:k,DEF:def,ALL:all}
	return nil
}

// Clears the rosters from the datastore
func ClearRosters(r *http.Request) (error) {
	c := appengine.NewContext(r)
	for _,p := range PLAYERS.ALL {
		key := datastore.NewKey(c, "player", p.PlayerID,0,nil)
		err := datastore.Delete(c,key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Retrieves the rosters from the datastore into the instance
func SyncRosters(r *http.Request) (error) {
	qb := make(map[string]Player)
	rb := make(map[string]Player)
	wr := make(map[string]Player)
	te := make(map[string]Player)
	k := make(map[string]Player)
	def := make(map[string]Player)
	all := make(map[string]Player)
	c := appengine.NewContext(r)
	q := datastore.NewQuery("player")
	//b := bytes.NewBuffer(nil)
	for t:= q.Run(c); ; {
		var p Player
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}
		id := p.PlayerID
		pos := p.Position
		switch pos {
		case "QB":
			qb[id] = p
		case "RB":
			rb[id] = p
		case "WR":
			wr[id] = p
		case "TE":
			te[id] = p
		case "K":
			k[id] = p
		case "DEF":
			def[id] = p
		}
		all[id] = p
	}
	PLAYERS = Players{QB:qb,RB:rb,WR:wr,TE:te,K:k,DEF:def,ALL:all}

	return nil
}

// Serves a pick and removes the picked player from the instance
func FindPlayer(player string, team, round int) {
	// When replacing a player make sure you put it back into the pool
	if _, present := PLAYERS.QB[player]; present {
		PICKS[team][round] = PLAYERS.QB[player]
		delete(PLAYERS.QB, player)
	}
	if _, present := PLAYERS.RB[player]; present {
		PICKS[team][round] = PLAYERS.RB[player]
		delete(PLAYERS.RB, player)
	}
	if _, present := PLAYERS.WR[player]; present {
		PICKS[team][round] = PLAYERS.WR[player]
		delete(PLAYERS.WR, player)
	}
	if _, present := PLAYERS.TE[player]; present {
		PICKS[team][round] = PLAYERS.TE[player]
		delete(PLAYERS.TE, player)
	}
	if _, present := PLAYERS.K[player]; present {
		PICKS[team][round] = PLAYERS.K[player]
		delete(PLAYERS.K, player)
	}
	if _, present := PLAYERS.DEF[player]; present {
		PICKS[team][round] = PLAYERS.DEF[player]
		delete(PLAYERS.DEF, player)
	}
}

// Combine all the postional rosters into one big roster list
func GetRosters() map[string]Player {
	p := make(map[string]Player)
	for _,v := range PLAYERS.QB {
		p[v.PlayerID] = v
	}
	for _,v := range PLAYERS.RB {
		p[v.PlayerID] = v
	}
	for _,v := range PLAYERS.WR {
		p[v.PlayerID] = v
	}
	for _,v := range PLAYERS.TE {
		p[v.PlayerID] = v
	}
	for _,v := range PLAYERS.K {
		p[v.PlayerID] = v
	}
	for _,v := range PLAYERS.DEF {
		p[v.PlayerID] = v
	}
	return p
}

// Combine all the picks into one single dimensional list
func GetPicksList() []Player {
	p := make([]Player,1)
	for i:=NUMTEAMS;i>0;i-- {
		for j:=NUMROUNDS;j>0;j-- {
			if PICKS[i][j].Name != "" {
				p = append(p,PICKS[i][j])
			}
		}
	}
	return p[1:]
}

// ===============================
// Handlers
// ===============================

// Test page set up for conducting various tests
func test(w http.ResponseWriter, r *http.Request) {
	_ = TEMPLATES.ExecuteTemplate(w,"test.html",PLAYERS)
}

// The home page or login page
// Enter username and password to log in
// Allow possiblity to enter keepers
func index(w http.ResponseWriter, r *http.Request) {
	/* LEAVE COMMENTED FOR TESTING
	//Get Cookie
	//cookie, err := r.Cookie("username")
	//If exists 
	if err == nil {
		//forward to lobby
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/lobby", http.StatusFound)
	}
	*/
	errT := TEMPLATES.ExecuteTemplate(w,"index.html",nil)
	if errT != nil {
		http.Error(w, errT.Error(), http.StatusInternalServerError)
	}
}

// Handle login
// Handles the login info and redirects user to either the keeper page or the lobby page.
func login(w http.ResponseWriter, r *http.Request) {
	// Get login info
	t := r.FormValue("teamname")
	p := r.FormValue("pwd")
	k := r.FormValue("keep")
	// If incorrect password
	if p != teamPassword[t] {
		// Relogin
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Set Cookie
	cookie := &http.Cookie{Name: "username", Value: t}
	http.SetCookie(w, cookie)
	// If they want keepers or not
	if k == "yes" {
		http.Redirect(w, r, "/keepers", http.StatusFound)
	} else {
		http.Redirect(w, r, "/lobby", http.StatusFound)
	}
}

// Research Page
// Search for a player and if found, display that player's recent news
// FUTURE: Display the player's stats/information
func research(w http.ResponseWriter, r *http.Request) {
	player := r.FormValue("player")
	if player != "" {
		// Get player reaserch stuff
		c := appengine.NewContext(r)
		client := urlfetch.Client(c)
		link := fmt.Sprintf("http://api.espn.com/v1/sports/football/nfl/athletes/%s/news?apikey=2n9w6hnjnjbeajwgd3bze9uz",player)
		res, err := client.Get(link)
		if err != nil {
			fmt.Fprintf(w, `<html>ERROR: %v <br /><a href="/lobby">Back</a></html>`,err)
			return
		}
		defer res.Body.Close()
		data := make([]byte,1e6)
		count,_ := io.ReadFull(res.Body,data)
		var rsrch Research
		err = json.Unmarshal(data[:count], &rsrch)
		if err != nil {
			fmt.Fprintf(w, `<html>ERROR: %v <br /><a href="/admin">Back</a></html>`,err)
			return
		}
		errT := TEMPLATES.ExecuteTemplate(w,"research.html",rsrch.Headlines[0])
		if errT != nil {
			http.Error(w, errT.Error(), http.StatusInternalServerError)
		}
	} else {
		errT := TEMPLATES.ExecuteTemplate(w,"research.html",nil)

		if errT != nil {
			http.Error(w, errT.Error(), http.StatusInternalServerError)
		}
	}
}

// News Page
// Display the most recent NFL news 
func news(w http.ResponseWriter, r *http.Request) {
	//developer.espn.com/docs
		c := appengine.NewContext(r)
		client := urlfetch.Client(c)
		// http://api.espn.com/:version/:resource/:method?apikey=:yourkey
		res, err := client.Get("http://api.espn.com/v1/sports/football/nfl/news/headlines?apikey=2n9w6hnjnjbeajwgd3bze9uz")
		if err != nil {
			fmt.Fprintf(w, `<html>ERROR: %v <br /><a href="/lobby">Back</a></html>`,err)
			return
		}
		defer res.Body.Close()
		data := make([]byte,1e6)
		count,_ := io.ReadFull(res.Body,data)
		var n News
		err = json.Unmarshal(data[:count], &n)
		if err != nil {
			fmt.Fprintf(w, `<html>ERROR: %v <br /><a href="/admin">Back</a></html>`,err)
			return
		}
	errT := TEMPLATES.ExecuteTemplate(w,"news.html",n)
	if errT != nil {
		http.Error(w, errT.Error(), http.StatusInternalServerError)
	}
}

// The lobby 
// Where the user's main draft experience occurs.
// Displays the user's picks as well as the entire league's recent picks
// Displays each`team's picks
// Allows user to make a pick
// FUTURE: Add a chat window (Identical to the Cuddle demo)
func lobby(w http.ResponseWriter, r *http.Request) {
	// Get Cookie
	cookie, err := r.Cookie("username")
	// If doesn't exist
	if err != nil {
		// Redirect to index
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}

	// Rerecord what picks the user has made up to this point
	var p [15]Player
	for pi:=0;pi<NUMROUNDS;pi++ {
		p[pi]=PICKS[teamNumber[cookie.Value]][pi+1]
	}

	// Create User using cookie username to lookup Name and Team
	u := User{Name: teamName[cookie.Value], Team: teamTeam[cookie.Value], Picks: p, Number: teamNumber[cookie.Value]}
	pause := ""
	if !PAUSE {
		pause = "pause"
	}
	page := &Page{League: TEAMS, User: u, Rosters: PLAYERS, AllPicks: ALLPICKS, Pause: pause}
	errT := TEMPLATES.ExecuteTemplate(w,"lobby.html",page)
	if errT != nil {
		http.Error(w, errT.Error(), http.StatusInternalServerError)
	}

	// Set cookie
	http.SetCookie(w, cookie)
}

// Handle Logout
// Logs the user out and redirects to the login page.
func logout(w http.ResponseWriter, r *http.Request) {
	// Get Cookie
	cookie, err := r.Cookie("username")
	// If it doesn't exist
	if err != nil {
		//redirect to index check if there is an http.StatusNotLoggedIn
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}
	// Delete cookie
	cookie.MaxAge = -1 //DOESN'T WORK????
	// Redirect to login
	http.Redirect(w, r, "/", http.StatusFound)
}

// Keeper Selection Page
// Inputs the keepers that the user wishes to claim.
func keepers(w http.ResponseWriter, r *http.Request) {
	// Get Cookie
	cookie, err := r.Cookie("username")
	// If doesn't exist
	if err != nil {
		// Redirect to login
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}

	// Create User using cookie username to lookup Name and Team
	u := User{Name: teamName[cookie.Value], Team: teamTeam[cookie.Value]}
	page := Page{User: u, Rosters: PLAYERS}
	errT := TEMPLATES.ExecuteTemplate(w,"keepers.html",page)
	if errT != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Reset Cookie
	http.SetCookie(w, cookie)
}

// Handle Keepers
// Submits the keepers as picks and forwards the user to the lobby.
func keepme(w http.ResponseWriter, r *http.Request) {
	//Get Cookie
	cookie, err := r.Cookie("username")
	//if doesn't exist
	if err != nil {
		//redirect to index check if there is an http.StatusNotLoggedIn
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}
	// Retrieve the form values
	var k1,k2,k3 string
	var r1,r2,r3 int
	num,_ := strconv.Atoi(r.FormValue("num"))
	k1 = r.FormValue("player1")
	r1,_ = strconv.Atoi(r.FormValue("round1"))
	if num > 1 {
		k2 = r.FormValue("player2")
		r2,_ = strconv.Atoi(r.FormValue("round2"))
		if num > 2 {
			k3 = r.FormValue("player3")
			r3,_ = strconv.Atoi(r.FormValue("round3"))
		}
	}
	// Use the cookie's value to lookup the teamNumber
	team := teamNumber[cookie.Value]
	// Add keepers
	FindPlayer(k1, team, r1)
	if num > 1 {
		FindPlayer(k2, team, r2)
		if num > 2 {
			FindPlayer(k3, team, r3)
		}
	}
	/*
	PICKS[team][r1] = k1
	PICKS[team][r2] = k2
	PICKS[team][r3] = k3
	*/
	// Redirect to the lobby
	http.Redirect(w, r, "/lobby", http.StatusFound)
}

// Handle Pick
// Submits the draft pick and redirects the user to the lobby.
func picked(w http.ResponseWriter, r *http.Request) {
	//Get Cookie
	cookie, err := r.Cookie("username")
	//if doesn't exist
	if err != nil {
		//redirect to index check if there is an http.StatusNotLoggedIn
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}
	// Use the cookie's value to lookup the teamNumber
	team := teamNumber[cookie.Value]
	// Retrieve the form values
	player := r.FormValue("player")
	position := r.FormValue("position")

	var pick Player
	team--; // Because TEAMS is 0 based
	if position == "qb" {
		pick = PLAYERS.QB[player]
		TEAMS[team].QB = append(TEAMS[team].QB,pick)
		delete(PLAYERS.QB, player)
	} else if position == "rb" {
		pick = PLAYERS.RB[player]
		TEAMS[team].RB = append(TEAMS[team].RB,pick)
		delete(PLAYERS.RB, player)
	} else if position == "wr" {
		pick = PLAYERS.WR[player]
		TEAMS[team].WR = append(TEAMS[team].WR,pick)
		delete(PLAYERS.WR, player)
	} else if position == "te" {
		pick = PLAYERS.TE[player]
		TEAMS[team].TE = append(TEAMS[team].TE,pick)
		delete(PLAYERS.TE, player)
	} else if position == "k" {
		pick = PLAYERS.K[player]
		TEAMS[team].K = append(TEAMS[team].K,pick)
		delete(PLAYERS.K, player)
	} else if position == "def" {
		pick = PLAYERS.DEF[player]
		TEAMS[team].DEF = append(TEAMS[team].DEF,pick)
		delete(PLAYERS.DEF, player)
	}
	ALLPICKS = append(ALLPICKS, pick)
	team++; // Because TEAMS is 0 based

	// Count how many picks you have made
	num := 0
	for i:=0;i<NUMROUNDS;i++ {
		if PICKS[team][i+1].Name != "" {
			num++
		} else {
			break
		}
	}
	// Append next pick
	if num < NUMROUNDS {
		PICKS[team][num+1] = pick
	}
	// Reset Cookie?
	//http.SetCookie(w, &cookie)
	// Redirect to the lobby
	http.Redirect(w, r, "/lobby", http.StatusFound)
}

// Timer page
// Displays a timer to be retrieved through javascript by other pages
func timer(w http.ResponseWriter, r *http.Request) {
	page := getTime()
	fmt.Fprint(w, page)
}


// Draft Board Page
// Here the user can see all of the draft picks.
// FUTURE: Use the App Engine Channel API
func draft(w http.ResponseWriter, r *http.Request) {
	// Generate HTML for the draft table
	page := `
<html>
<head>
<link rel="stylesheet" type="text/css" href="stylesheets/style.css" />
`
	if !PAUSE {
		page += `<script type="text/javascript" src="javascripts/timer.js"></script>`
	}
	page += `
</head>
<body>
	<div class="navbar">
		<div class="navbar-inner">
			<a href="/" class="brand">
				<img src="/images/title.png" />
			</a>
			<div id="timer">
`
	page += getTime()
	page +=
`
			</div>
			<ul class="nav">
				<li><a href="/lobby">Lobby</a></li>
				<li><a href="/draft">Draft Board</a></li>
				<li><a href="/admin">Admin</a></li>
				<li><a href="/news">News</a></li>
				<li><a href="/research">Research</a></li>
				<li><a href="/logout">Logout</a></li>
			</ul>
		</div>
	</div>
	<table class="draft" table-layout="fixed">
`
	page += `		<tr>\n`
	for i:=0;i<=NUMTEAMS;i++ {
		if i==0 {
			page += "			<th>"
			page += "</th>\n"
		} else {
			page += "			<th>"
			page += teamTeam[rlookup(teamNumber,i)] //Team name headers
			page += "</th>\n"
		}
	}
	page += "		</tr>\n"
	for i:=1;i<=NUMROUNDS;i++ {
		page += "		<tr>\n"
		for j:=0;j<=NUMTEAMS;j++ {
			if j==0 {
				page += "			<td>"
				page += strconv.Itoa(i) // Round numbers
				page += "</td>\n"
			} else {
				page += "			<td>"
				page += PICKS[j][i].Name // Each pick
				page += " "
				page += PICKS[j][i].Team // Each pick
				page += "</td>\n"
			}
		}
		page += "		</tr>\n"
	}
	page += `
</table>

<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.3.0/jquery.min.js"></script>
</html>`

	fmt.Fprint(w, page)
}

// The admin page 
// Provides special settings including the ability to override picks and manage rosters
func admin(w http.ResponseWriter, r *http.Request) {
	// Get Cookie
	cookie, err := r.Cookie("username")
	// If doesn't exist
	if err != nil {
		// Redirect to index
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}
	// Check admin permissions
	if cookie.Value != ADMINS[0] && cookie.Value != ADMINS[1] {
		http.Redirect(w, r, "/lobby", http.StatusForbidden)
		return
	}

	PLAYERS.ALL = GetRosters()
	pause := ""
	if !PAUSE {
		pause = "pause"
	}
	page := Page{Pause: pause, Players: PLAYERS}
	errT := TEMPLATES.ExecuteTemplate(w,"admin.html",page)
	if errT != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Set cookie
	//http.SetCookie(w, cookie)
}


func setadmin(w http.ResponseWriter, r *http.Request) {
	// Get which function to perform
	adminfunction := r.FormValue("admin")

	if adminfunction == "reset" {
		// Reset Clock
		CLOCK = time.Now()
		// Reset all draft picks
		var p Player
		for i:=0;i<NUMTEAMS;i++ {
			for j:=0;j<NUMROUNDS;j++ {
				PICKS[i][j] = p
			}
		}
		// FUTURE: Make sure EVERYTHING is cleared.
		SyncRosters(r)
	} else if adminfunction == "override" {
		// Override Pick
		// Get form values
		team := r.FormValue("team")
		round,_ := strconv.Atoi(r.FormValue("round"))
		player := r.FormValue("player")
		// Put the previous pick back in the pool
		oldpick := PICKS[teamNumber[team]][round]
		oldpos := oldpick.Position
		oldid := oldpick.PlayerID
		if oldpos == "QB" {
			PLAYERS.QB[oldid] = oldpick
		}
		if oldpos == "RB" {
			PLAYERS.RB[oldid] = oldpick
		}
		if oldpos == "WR" {
			PLAYERS.WR[oldid] = oldpick
		}
		if oldpos == "TE" {
			PLAYERS.TE[oldid] = oldpick
		}
		if oldpos == "K" {
			PLAYERS.K[oldid] = oldpick
		}
		if oldpos == "DEF" {
			PLAYERS.DEF[oldid] = oldpick
		}

		// Override draft pick
		if _, present := PLAYERS.QB[player]; present {
			PICKS[teamNumber[team]][round] = PLAYERS.QB[player]
			delete(PLAYERS.QB, player)
		}
		if _, present := PLAYERS.RB[player]; present {
			PICKS[teamNumber[team]][round] = PLAYERS.RB[player]
			delete(PLAYERS.RB, player)
		}
		if _, present := PLAYERS.WR[player]; present {
			PICKS[teamNumber[team]][round] = PLAYERS.WR[player]
			delete(PLAYERS.WR, player)
		}
		if _, present := PLAYERS.TE[player]; present {
			PICKS[teamNumber[team]][round] = PLAYERS.TE[player]
			delete(PLAYERS.TE, player)
		}
		if _, present := PLAYERS.K[player]; present {
			PICKS[teamNumber[team]][round] = PLAYERS.K[player]
			delete(PLAYERS.K, player)
		}
		if _, present := PLAYERS.DEF[player]; present {
			PICKS[teamNumber[team]][round] = PLAYERS.DEF[player]
			delete(PLAYERS.DEF, player)
		}
		ALLPICKS = append(ALLPICKS, PLAYERS.ALL[player])
	} else if adminfunction == "start" {
		// Start Clock
		PAUSE = false
		CLOCK = time.Now()
	} else if adminfunction == "stop" {
		// Stop Clock
		PAUSE = true
	} else if adminfunction == "rosters" {
		// Update Rosters
		c := appengine.NewContext(r)
		client := urlfetch.Client(c)
		res, err := client.Get("http://api.fantasyfootballnerd.com/ffnPlayersXML.php?apiKey=2012050338875903")
		//res, err := client.Get("http://squinn.php.cs.dixie.edu/players.xml") // For testing only
		if err != nil {
			fmt.Fprintf(w, `<html>ERROR: %v <br /><a href="/admin">Back</a></html>`,err)
			return
		}
		defer res.Body.Close()
		data := make([]byte,1e6)
		var plist []Player
		ffn := &FFN{Results: plist}
		count,_ := io.ReadFull(res.Body,data)
		err = xml.Unmarshal(data[:count], ffn)
		if err != nil {
			fmt.Fprintf(w, `<html>ERROR: %v <br /><a href="/admin">Back</a></html>`,err)
			return
		}
		err = UpdateRosters(ffn.Results,r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if adminfunction == "sync" {
		SyncRosters(r)
	} else if adminfunction == "clear" {
		ClearRosters(r)
	}
	// Redirect to draft board
	http.Redirect(w, r, "/lobby", http.StatusFound)
}

func init() {

	NUMROUNDS = 15
	NUMTEAMS = 12
	PAUSE = true

	for i:=1;i<=NUMTEAMS;i++ {
		t := rlookup(teamNumber,i)
		id := strconv.Itoa(i)
		TEAMS = append(TEAMS,Team{Name: teamTeam[t],Number:i,TabID:"tabs2-"+id})
	}

	http.HandleFunc("/test", test)

	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)

	http.HandleFunc("/news", news)
	http.HandleFunc("/research", research)

	http.HandleFunc("/keepers", keepers)
	http.HandleFunc("/keepme", keepme)

	http.HandleFunc("/lobby", lobby)
	http.HandleFunc("/picked", picked)

	http.HandleFunc("/draft", draft)
	http.HandleFunc("/timer", timer)

	http.HandleFunc("/admin", admin)
	http.HandleFunc("/setadmin", setadmin)

	http.HandleFunc("/logout", logout)
	http.ListenAndServe(":8080", nil)
}
