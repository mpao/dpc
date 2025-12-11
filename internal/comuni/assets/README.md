# Comuni

La lista dei comuni è generata a partire dai dati
ricavabili da https://github.com/opendatasicilia/comuni-italiani

## 1. Problemi

Sia le allerte, che i dati meteo hanno encoding errato. Questo vuol dire che scaricando un json troverai i nomi dei comuni con lettere accentate scritti in maniera errata.

Secondo problema: i comuni possono trovarsi in zone di allarme differenti, quindi ripetuti, oppure avere zone definite con nomi
leggermente diversi di volta in volta.

Terzo problema: sempre nei json, vengono utilizzati comuni pre 2015. Questo comporta allarmi su comuni che hanno cambiato nome o la mancanza di allarmi sui nuovi comuni. Ad esempio

```json
	"024126": {
		"Comune": "Colceresa",
		"Alias": ["Mason Vicentino", "Molvena"],
		"ID": "024126",
		"Zone": [
			"Alto Brenta-Bacchiglione-Alpone",
			"Trentino e Alto piano dei sette comuni"
		],
		"Prov": "Vicenza",
		"Sigla": "VI",
		"Reg": "Veneto",
		"Info": "http://comune.colceresa.vi.it",
		"Lat": 45.718116,
		"Lon": 11.607036,
		"Pop": 5986
	},
```

Il comune di Colceresa è nato nel 2019 dalla fusione di Molvena e Mason Vicentino (fonte https://it.wikipedia.org/wiki/Colceresa)

Per risolvere il secondo e terzo problema utilizza `Alias` e `Zone`

## 2. Omonimie

Esistono una manciata di omonimie a cui prestare attenzione. 

|id|comune|provincia|
|--|------|---------|
|022035|Calliano|TN|
|005014|Calliano|TO|
|016065|Castro|BG|
|075096|Castro|LE|
|013130|Livo|CO|
|022106|Livo|TN|
|041041|Peglio|PU|
|013178|Peglio|CO|
|022165|Samone|TN|
|001235|Samone|TO|
|083090|San Teodoro|ME|
|090092|San Teodoro|SS|


## 3. Soluzione

Per avere una corrispondenza evento => comune viene generato un ID a partire dal nome del comune e la sua zona di allerta. I comuni con alias o più zone di appartenenza avranno un intero keyring contenente tutte le possibili combinazioni di nome/zona
