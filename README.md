## Composer Update Bot
https://github.com/vektorprogrammet/github-bot/blob/master/bot/composer-update-bot.go

Composer Update Bot oppdaterer alle composer dependencies i et gitt repository. Oppdateringen kjøres i et CRON-intervall (hver søndag kl. 20):
```go
startUpdateBot("git@github.com:vektorprogrammet/vektorprogrammet.git", "0 0 20 * * SUN")
```

## Code Style Bot
https://github.com/vektorprogrammet/github-bot/blob/master/bot/code-style-bot.go

Når det opprettes en PR til https://github.com/vektorprogrammet/vektorprogrammet vil Code Style Bot kjøre [PHP Coding Standards Fixer](https://github.com/FriendsOfPHP/PHP-CS-Fixer) på brancen til PRen. Hvis den oppdager feil i code style vil den rette opp feilen og commite fiksen til PRen.

## DB Migration Bot
https://github.com/vektorprogrammet/github-bot/blob/master/bot/db-migration-bot.go

Når det opprettes en PR til https://github.com/vektorprogrammet/vektorprogrammet vil DB Migration Bot sjekke om det trengs å gjøre endringer i databaseskjemaet for den nye koden. Hvis kodeendringene vil kreve endringer i databaseskjemaet, og en migration ikke allerede er opprettet, vil DB Migration Bot lage et migration script og commite det til PRen.
