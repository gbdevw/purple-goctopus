package websocket

import "github.com/stretchr/testify/suite"

/*************************************************************************************************/
/* UNIT TEST SUITE                                                                               */
/*************************************************************************************************/

type KrakenSpotPublicWebsocketClientTestSuite struct {
	suite.Suite
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the regexe matches the event type of all message types which can be recieved by the client.
func (suite *KrakenSpotPublicWebsocketClientTestSuite) TestMessageTypeRegexMatching() {

}
