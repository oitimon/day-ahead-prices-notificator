To create a script that does next steps:

1. Downloads JSON from https://api.com/v1/energyprices?fromDate=2025-02-27T23%3A00%3A00.000Z&tillDate=2025-02-28T22%3A59%3A59.000Z&interval=4&usageType=1&inclBtw=false
where fromDate must be start of the next day (tomorrow) from now in Amsterdam timezone (when fromDate and tillDate are converted to UTC).
inclBtw parameter must be defined in the environment variable.
The JSON structure is:
"""JSON
   {
       "Prices": [
         {
            "price": 0.0971,
            "readingDate": "2025-02-27T23:00:00Z"
         },
         {
            "price": 0.11054,
            "readingDate": "2025-02-28T00:00:00Z"
         }
       ]
   }
"""
where there are usually 24 elements in the array (one for each hour of the day).
Hostname of the URL must be stored in the environment variable.

2. Creates a bar chart with the prices for the next day (from 00:00 to 23:00 in Amsterdam time) and saves it to a graphic file. The chart must have agenda and price values for every hour.

3. Send a message to Telegram bot (token and chatId must be stored in the environment variables) with the chart attached.
The message starts with the text "EPEX NL DA 2025-02-28", where "2025-02-28" is the date of the prices we're working on.
If there are some prices equal or higher than HIGH_PRICE (defined in env) or equal or lower than LOW_PRICE (defined in env), the message must contain the text "There are High prices" or "There are Low prices" or "There are High/Low prices" (when both are present) respectively.
The chart has to be deleted after sending the message (even if error happened).


Conditions:
* If some of the steps fail, the script must send message to the Telegram bot "Error for 2025-02-28", where "2025-02-28" is the date of the prices we're working on.
* If endpoint in step 1 doesn't return prices and has a structure like
"""JSON
{
  "Prices": [],
}
"""
the script must send message to the Telegram bot "No prices for 2025-02-28", where "2025-02-28" is the date of the prices we're working on.
* github.com/kelseyhightower/envconfig must be used for environment variables (so we'll have some Config structure).
* every .go file must have _test.go file with tests for the functions.
