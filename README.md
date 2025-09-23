A Go-based demonstration of multi-agents capabilities using Claude 3.5 Sonnet via OpenRouter. The agent should be able to answser the questions like, by using two sub-agents. One of them is handling the temperature and the other is foucsing on datetime. Both of them should use external MCP servers implemented to get the temperatur or datetime data. 

- What is the temperature in New York City right now?
- What is the datetime of New York City now?
- What is the datetime and temperature of New York City now?

Supposedly, if the user is asking for both datetime and temperature, the agent can use both datetime and temperature sub-agents in parallel.

## 📋 Prerequisites

- Go 1.21+ (tested with 1.25.1)
- Use https://github.com/nlpodyssey/openai-agents-go as the AI Agents development SDK
- Use https://github.com/modelcontextprotocol/go-sdk for the MCP implemenation
- OpenRouter API key for Claude access
- No additional API keys required (weather and datetime are free)

