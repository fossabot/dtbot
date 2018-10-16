package cmd

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"../bot"
)

// YoutubeCommand youtube handler
func YoutubeCommand(ctx bot.Context) {
	sess := ctx.Sessions.GetByGuild(ctx.Guild.ID)
	if len(ctx.Args) == 0 {
		ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_no_args"))
		return
	}
	switch ctx.Args[0] {
	case "play":
		if sess == nil {
			ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("player_not_in_voice"))
			return
		}
		queue := sess.Queue
		if !queue.HasNext() {
			ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_queue_is_empty"))
			return
		}
		cmsg := ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_starting"))
		go queue.Start(sess, cmsg, func(msg string, cmsg *discordgo.Message) {
			switch msg {
			case "stop":
				ctx.EditEmbed(cmsg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_stopped"), true)
				break
			case "finish":
				ctx.EditEmbed(cmsg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_finished"), true)
				break
			default:
				ctx.EditEmbed(cmsg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), fmt.Sprintf("%v: %v", ctx.Loc("youtube_now_playing"), msg), true)
			}
		})
	case "stop":
		if sess == nil {
			ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("player_not_in_voice"))
			return
		}
		if sess.Queue.HasNext() {
			sess.Queue.Clear()
		}
		sess.Stop()
		ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_stopped"))
	case "add":
		newargs := ctx.Args[1:]
		if len(newargs) == 0 {
			ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_args_missing"))
			return
		}
		if sess == nil {
			ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("player_not_in_voice"))
			return
		}
		msg := ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_adding_song"))
		for _, arg := range newargs {
			t, inp, err := ctx.Youtube.Get(arg)

			if err != nil {
				ctx.EditEmbed(msg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("error"), true)
				fmt.Println("error getting input,", err)
				return
			}

			switch t {
			case bot.ERROR_TYPE:
				ctx.EditEmbed(msg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("error"), true)
				fmt.Println("error type", t)
				return
			case bot.VIDEO_TYPE:
				{
					video, err := ctx.Youtube.Video(*inp)
					if err != nil {
						ctx.EditEmbed(msg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("error"), true)
						fmt.Println("error getting video1,", err)
						return
					}
					song := bot.NewSong(video.Media, video.Title, arg)
					sess.Queue.Add(song)
					ctx.EditEmbed(msg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), fmt.Sprintf(ctx.Loc("youtube_added_format"), song.Title), true)
					break
				}
			case bot.PLAYLIST_TYPE:
				{
					videos, err := ctx.Youtube.Playlist(*inp)
					if err != nil {
						ctx.EditEmbed(msg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("error"), true)
						fmt.Println("error getting playlist,", err)
						return
					}
					for _, v := range *videos {
						id := v.Id
						_, i, err := ctx.Youtube.Get(id)
						if err != nil {
							ctx.EditEmbed(msg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("error"), true)
							fmt.Println("error getting video2,", err)
							continue
						}
						video, err := ctx.Youtube.Video(*i)
						if err != nil {
							ctx.EditEmbed(msg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("error"), true)
							fmt.Println("error getting video3,", err)
							return
						}
						song := bot.NewSong(video.Media, video.Title, arg)
						sess.Queue.Add(song)
					}
					ctx.EditEmbed(msg.ID, fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_added"), true)
					break
				}
			}
		}
	case "clear":

		if sess == nil {
			ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("player_not_in_voice"))
			return
		}
		if !sess.Queue.HasNext() {
			ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_queue_is_empty"))
			return
		}
		sess.Queue.Clear()
		ctx.ReplyEmbed(fmt.Sprintf("%v:", ctx.Loc("youtube")), ctx.Loc("youtube_queue_cleared"))
	}
}