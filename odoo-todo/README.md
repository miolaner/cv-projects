# Todo App - Odoo-moduuli

Yksinkertainen tehtävienhallintasovellus Odoo-järjestelmälle. Sovellus on toteutettu Odoo 16.0 -versiolle.

## Ominaisuudet

### Tehtävien hallinta
- Tehtävien luonti, muokkaus ja poisto
- Tehtävien arkistointi (pehmeä poisto)
  - Arkistoidut tehtävät piilottuvat oletuksena näkymistä
  - Arkistoidut tehtävät voi palauttaa myöhemmin
  - Nähtävissä "Arkistoidut"-suodattimella
- Tehtävien pysyvä poisto mahdollista (tietokantatason poisto)
- Tehtävän tiedot:
  - Nimi
  - Kuvaus
  - Aloitusaika
  - Lopetusaika
  - Käytetty aika (lasketaan automaattisesti)
  - Vastuuhenkilö
  - Arkistointitila (aktiivinen/arkistoitu)

### Tehtävien tilat
1. Luonnos (Draft)
2. Työn alla (In Progress)
3. Valmis (Done)

### Aikaseuranta
- Automaattinen ajan laskenta
- Aloitus- ja lopetusaikojen tallennus
- Käytetyn ajan näyttö tunteina

### Käyttöliittymä
- Lista-, lomake- ja kanban-näkymät
- Monipuoliset hakutoiminnot
- Ryhmittely tilan ja käyttäjän mukaan
- Arkistointimahdollisuus

### Lisäominaisuudet
- Käyttäjien viestintä ja kommentointi (Chatter)
- Tehtävien seuranta
- Aktiviteettien hallinta

## Tekninen toteutus

### Rakenne
```
todo_app/
├── __init__.py
├── __manifest__.py
├── models/
│   ├── __init__.py
│   └── todo_task.py
├── views/
│   ├── todo_task_views.xml
│   └── menu_views.xml
└── security/
    └── ir.model.access.csv
```

### Tietomalli (todo.task)
- name: Tehtävän nimi (Char)
- description: Kuvaus (Text)
- state: Tila (Selection)
- start_time: Aloitusaika (Datetime)
- end_time: Lopetusaika (Datetime)
- time_spent: Käytetty aika (Float, laskettu kenttä)
- user_id: Vastuuhenkilö (Many2one -> res.users)
- active: Arkistointitila (Boolean)

### Docker-käyttöönotto
1. Asenna Docker ja Docker Compose
2. Kloonaa repositorio
3. Käynnistä kontit:
   ```bash
   docker-compose up -d
   ```
4. Avaa selain osoitteessa http://localhost:8069
5. Luo uusi tietokanta ja asenna Todo App -moduuli

### Käyttöoikeudet
- Peruskäyttäjät voivat:
  - Luoda uusia tehtäviä
  - Muokata omia tehtäviään
  - Nähdä kaikki tehtävät
  - Seurata tehtäviä ja kommentoida
  - Arkistoida omia tehtäviään
  - Poistaa omia tehtäviään pysyvästi
  - Palauttaa arkistoituja tehtäviä

## Kehitysympäristö
- Python 3.x
- Odoo 16.0
- PostgreSQL
- Docker & Docker Compose

## Asennus kehitysympäristöön
1. Kloonaa repositorio Odoo-addons-kansioon
2. Asenna riippuvuudet:
   ```bash
   pip install -r requirements.txt
   ```
3. Lisää moduuli Odoon conf-tiedostoon:
   ```
   addons_path = /path/to/addons,/path/to/todo_app
   ```
4. Käynnistä Odoo-palvelin
5. Asenna moduuli Odoon käyttöliittymästä

## Lisenssi
MIT License

## Tuki
Ongelmatilanteet ja kysymykset voi raportoida Issues-osiossa.
