package quizduell

type GameState int8

const (
	Waiting GameState = iota
	Active
	Finished
	_
	_
	GivenUp
)

func (s GameState) String() string {
	switch s {
	case Waiting:
		return "Waiting"
	case Active:
		return "Active"
	case Finished:
		return "Finished"
	case GivenUp:
		return "GivenUp"
	default:
		panic("invalid value for GameState")
	}
}

type User struct {
	AvatarCode           string `json:"avatar_code"`
	Name                 string `json:"name"`
	ID                   int    `json:"user_id,string"`
	TT                   string `json:"tt"`
	QC                   bool   `json:"qc"`
	QuestionReviewer     int    `json:"q_reviewer"`
	NumApprovedQuestions int    `json:"n_approved_questions"`
	Key                  int    `json:"key"`
	Rating               int    `json:"rating"`
	Friends              []User `json:"friends"`
	FacebookID           int    `json:"facebook_id,string"`
	Games                []Game `json:"games"`
	Email                string `json:"email"`
}

type Game struct {
	ID              int             `json:"game_id"`
	CategoryChoices []int           `json:"cat_choices"`
	ElapsedMinutes  int             `json:"elapsed_min"`
	Messages        []InGameMessage `json:"messages"`
	Opponent        User            `json:"opponent"`
	OpponentAnswers []int           `json:"opponent_answers"`
	YourAnswers     []int           `json:"your_answers"`
	YourTurn        bool            `json:"your_turn"`
	Questions       []Question      `json:"questions"`
	RatingBonus     int             `json:"rating_bonus"`
	GameState       `json:"state"`
}

type Question struct {
	QuestionText string `json:"question"`
	Correct      string `json:"correct"`
	Wrong1       string `json:"wrong1"`
	Wrong2       string `json:"wrong2"`
	Wrong3       string `json:"wrong3"`
	Timestamp    string `json:"timestamp"`
	CategoryName string `json:"cat_name"`
	CategoryID   int    `json:"cat_id"`
	QuestionID   int    `json:"q_id"`
}

type GameStatistic struct {
	AvatarCode string `json:"avatar_code"`
	GamesLost  int    `json:"n_games_lost"`
	GamesTied  int    `json:"n_games_tied"`
	GamesWon   int    `json:"n_games_won"`
	Name       string `json:"name"`
	UserID     int    `json:"user_id,string"`
}

type InGameMessage struct {
	CreatedAt string `json:"created_at"`
	From      int    `json:"from"`
	ID        string `json:"id"`
	Text      string `json:"text"`
	To        int    `json:"to"`
}

type Popup struct {
	PopupMessage string `json:"popup_mess"`
	PopupTitle   string `json:"popup_title"`
}

type Status struct {
	LoggedIn bool `json:"logged_in"`
	*User    `json:"user"`
	Settings *struct {
		MaxFreeGames     int     `json:"max_free_games"`
		GiveUpPointLoss  int     `json:"give_up_point_loss"`
		AdProvider       string  `json:"ad_provider"`
		AdmobMedID       string  `json:"ad_mob_med_id"`
		AdmobMedSplashID string  `json:"admob_med_splash_id"`
		Fulmium          bool    `json:"fulmium"`
		Feo              bool    `json:"feo"`
		Feos             float64 `json:"feos"`
		PPF              float64 `json:"ppf"`
		CheckLimboGames  bool    `json:"check_limbo_games"`
		RefreshTableFreq int     `json:"refresh_table_freq"`
		RefreshFreq      int     `json:"refresh_freq"`
		SplashFreq       float64 `json:"splash_freq"`
		override         *struct {
			GameGiveUpMessage string `json:"GAME_GIVEUP_MESS"`
			InviteViaWhatsApp string `json:"INVITE_VIA_WHATSAPP"`
		}
	} `json:"settings"`
}

type CategoryStatistic struct {
	Percent      float64 `json:"percent"`
	CategoryName string  `json:"cat_name"`
}

type UserCategoryStatistics struct {
	CategoryStatistics []CategoryStatistic `json:"cat_stats"`
	GamesLost          int                 `json:"n_games_lost"`
	GamesPlayed        int                 `json:"n_games_played"`
	GamesTied          int                 `json:"n_games_tied"`
	GamesWon           int                 `json:"n_games_won"`
	PerfectGames       int                 `json:"n_perfect_games"`
	QuestionsAnswered  int                 `json:"n_questions_answered"`
	QuestionsCorrect   int                 `json:"n_questions_correct"`
	NumUsers           int                 `json:"n_users"`
	Rank               int                 `json:"rank"`
	Rating             int                 `json:"rating"`
}

type message struct {
	T bool  `json:"t"`
	U *User `json:"u"`

	Blocked        []User          `json:"blocked"`
	Users          []User          `json:"users"`
	Categories     map[int]string  `json:"cats"`
	GameStatistics []GameStatistic `json:"game_stats"`

	*Game          `json:"game"`
	*InGameMessage `json:"m"`
	*UserCategoryStatistics
	*Status
	*Popup

	RemovedID int  `json:"removed_id,string"`
	Access    bool `json:"access"`
}
