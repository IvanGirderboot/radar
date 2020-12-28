package main

import (
	"math/rand"
	"time"
)

func randomRadarQuote() string {
	quotes := []string{
		"These are the forms to get the forms to order more forms, sir.",
		"Here's a mover and a groover and it ain't by Herbert Hoover. It's for all you animals and music lovers.",
		"(Radar, seeing Klinger in pants) \"Don't I know your sister?\"",
		"Dear Mrs. Burns, I regret to inform that your husband has been seen out of uniform, and maybe you would like to know with who.",
		"Testing, 1,2,3,4,5,6,7,8 testing. A, B, C, D, E, F, G, H, I got a gal in Kalamazoo...",
		"I'm afraid he's doing some very important sleeping for the army right now.",
		"Why don't you sirs act like sirs, sir?",
		"Are you going to be a mother, sir?",
		"If I don't eat regularly, everything solid in my body turns to liquid.",
		"Oh, I am fine. Well, not really, I am closer to lousy than fine.",
		"Get away from me before I get physically emotional!",
		"What? He changed to psychiatry? That's crazy!",
		"Poetry, right? That's great how they can rhyme and be hot at the same time. ",
		"It's Mrs. Colonel, your wife, sir. ",
		"I've never seen you in your underneath before.",
		"If you want a drink, sir, -- compliments Henry Blake -- brandy, scotch, vodka. And for your convenience, all in the same bottle.",
		"As usual, I'm writing slowly because I know you can't read fast.",
		"Well, I guess that's a bear we all gotta cross.",
		"Testing, tes...1,2,3. Testing, 1, 2. Radar here, uh..there's nobody on the radio now except 'Seoul City' Sue so I figured I'd keep you entertained by reading you a letter from my mom. Here it goes. Dear Son, I got your lovely letter. You certainly asked a lot of questions. About the car, you may. About Jennifer next door, yes. About Eleanor Simon, she did once or twice but not too much. About your uncle Albert, uh no on drinking, yes on AA. About the dog Leon, three times in the bedroom, once under the washer, and twice on the cat. Testing, testing. About the cat, we don't have one anymore. About your cousin Ernie, he's in the...(explosion) Oh! Oh! Here we go again! Watch out!",
		"She kicked me and then she messed all my files from M to Zee and everything... And then she got mad.",
		"Listen, buddy, we're a hospital! How would you like it if we fired patients at you??",
		"I can't hear you! Boy, you've got the war on loud there!",
		"My father didn't have me til he was sixty-three. First time we played peek-a-boo he had a stroke.",
		"My bear went off!",
		"I've looked everywhere except the nurses's showers. Oh no, sir, I couldn't look in there - there might be naked female personnel showering with their clothes off!",
		"I don't think that this place is turning out to be that great an experience for me. I mean, I work under terrible pressure, and there's lots of death and destruction and stuff, but other than that I don't think I'm getting much out of it.",
		"When my Uncle Ed came home from World War I, his mother could tell from the look in his eyes that he hadn't been a good boy in France. She cried for three days. I just know when I get home, my mother's going to look at me and chuckle for a week.",
		"I'm the only one who's gonna leave this place younger than I was when I came in!",
		"Where were you originally born? I mean, as a child.",
		"I cleaned it for two hours! There was another mess under it!",
	}

	min := 0
	max := len(quotes) - 1
	rand.Seed(time.Now().UnixNano())
	q := rand.Intn(max-min) + min

	return quotes[q]

}
