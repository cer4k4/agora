package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	"github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
	"github.com/gin-gonic/gin"
)

var appID , appCertificate string

func main() {
	api := gin.Default()
	appID = "02db9b4268e14428aac4927db4b1d539"
	appCertificate = "cf7a3043794c46a0b07267d6cdc3ccd4"

	api.GET("/ping",func(c *gin.Context){
		c.JSON(200,gin.H{
			"message":"pong",
		})
	})
	api.GET("rtc/channelName/:role/:tokentype/:uid/", getRtcToken)
	api.GET("rtm/:uid/", getRtmToken)
	api.GET("rte/channelName/:role/:tokentype/:uid/",getBothTokens)
	api.Run(":8080")
}

func getBothTokens(c *gin.Context){
	log.Print("Dual Tokens\n")
	channelName,tokentype,uidStr,role,expireTimestamp,rtcParamerr:= parseRtcParams(c)

	if rtcParamerr != nil {
		c.Error(rtcParamerr)
		c.AbortWithStatusJSON(400,gin.H{
			"message":"Error To parse Param RTC token on dual token" + rtcParamerr.Error(),
			"status":400,
		})
		return
	}
	rtcToken,rtcTokenErr := generateRtcToken(channelName,uidStr,tokentype,role,expireTimestamp)
	rtmToken,rtmTokenErr := rtmtokenbuilder.BuildToken(appID,appCertificate,uidStr,rtmtokenbuilder.RoleRtmUser,expireTimestamp)

	if rtcTokenErr != nil {
		log.Println(rtcTokenErr)
		c.Error(rtcTokenErr)
		errMsg := "Error To Generate To RTC Token -" + rtcTokenErr.Error()
		c.AbortWithStatusJSON(400,gin.H{
			"status":400,
			"error":errMsg,
		})
	} else if rtmTokenErr != nil {
		log.Println(rtmTokenErr)
		c.Error(rtmTokenErr)
		errMsg := "Error To Generate To RTC Token -" + rtmTokenErr.Error()
		c.AbortWithStatusJSON(400,gin.H{
			"status":400,
			"error":errMsg,
		})
	} else {
		log.Panicln("Doual Tokens generated")
		c.JSON(200,gin.H{
			"rtcToken":rtcToken,
			"rtmToken":rtmToken,
		})
	}

}

func getRtmToken(c *gin.Context){
	log.Printf("rtm token\n")
	uidStr , expireTimestamp, err := parseRtmParams(c)

	if err != nil{
		c.Error(err)
		c.AbortWithStatusJSON(400,gin.H{
			"message":"Error Generating RTM Token: " + err.Error(),
			"status": 400,
		})
		return
	}
	rtmToken,tokenErr := rtmtokenbuilder.BuildToken(appID,appCertificate,uidStr,rtmtokenbuilder.RoleRtmUser,expireTimestamp)
	if tokenErr != nil {
		log.Println(err)
		c.Error(err)
		errMsg := "Error Generating RTM Token: " + tokenErr.Error()
		c.AbortWithStatusJSON(400,gin.H{
			"error":errMsg,
			"status":400,
		})
	} else {
		log.Println("RTM Token Generated")
		c.JSON(200,gin.H{
			"rtmToken": rtmToken,
		})
	}
}

func generateRtcToken(channelName,uidStr,tokentype string,role rtctokenbuilder.Role,expireTimestamp uint32) (rtcToken string,err error){

	if tokentype == "userAccount" {
		log.Printf("Building Token with userAccount:%s\n",uidStr)
		rtcToken,err = rtctokenbuilder.BuildTokenWithUserAccount(appID,appCertificate,channelName,uidStr,role,expireTimestamp)
		return rtcToken,err
	} else if tokentype == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr,10,64)
		if parseErr != nil{
			err = fmt.Errorf("failed to parse uidStr:%s to uint causing error:%s",uidStr,parseErr)
			return "",err
		}
		uid := uint32(uid64)
		log.Printf("Buildin Token with uid:%d\n", uid)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUID(appID,appCertificate,channelName,uid,role,expireTimestamp)
		return rtcToken, err
	} else {
		err = fmt.Errorf("failed to generate RTC token for unknow tokentype:%s",tokentype)
		log.Println(err)
		return "",err
	}


}

func getRtcToken(c *gin.Context) {
	log.Printf("rtc token\n")
	channelName,tokentype,uidStr,role,expireTimestamp,err := parseRtcParams(c)
	if err != nil{
		c.Error(err)
		c.AbortWithStatusJSON(400,gin.H{
			"message":"Error Generating RTC token: " + err.Error(),
			"status":400,
		})
		return
	}
	rtcToken,tokenErr := generateRtcToken(channelName,uidStr,tokentype,role,expireTimestamp)

	if tokenErr != nil {
		log.Println(tokenErr)
		c.Error(tokenErr)
		errMsg := "Error Generating RTC Token - " + tokenErr.Error()
		c.AbortWithStatusJSON(400,gin.H{
			"status":400,
			"error":errMsg,
		})
	} else {
		log.Println("Hey AkA RTC Token Generated")
		c.JSON(200,gin.H{
			"rtcToken": rtcToken,
		})
	}
}

func parseRtcParams(c *gin.Context) (channelName,tokentype,uidStr string,role rtctokenbuilder.Role,expireTimestamp uint32,err error) {
	channelName = c.Param("channelName")
	roleStr := c.Param("role")
	tokentype = c.Param("tokentype")
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry","3600")

	if roleStr == "publisher"{
		role = rtctokenbuilder.RolePublisher
	} else {
		role = rtctokenbuilder.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime,10,64)

	if parseErr != nil{
		err = fmt.Errorf("failed to parse expireTime:%s, cusing error:%s ",expireTime,parseErr)
	}
	expireTimeSecound := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeSecound

	return channelName,tokentype,uidStr,role,expireTimestamp,err
}




func parseRtmParams(c *gin.Context) (uidStr string,expireTimestamp uint32,err error) {
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry","3600")

	expireTime64,parseErr := strconv.ParseUint(expireTime,10,64)
	if parseErr != nil {
		err = fmt.Errorf("failed to expireTime:%s, Cusing error:%s",expireTime,parseErr)
	}

	expireTimeSecound := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeSecound

	return uidStr,expireTimestamp,err
}


