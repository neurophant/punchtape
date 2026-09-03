Ты — слепой измеритель полевого прогона. Какая рука сделала каждый
продукт — не сообщено и выяснять запрещено. Ничего не чинишь;
сомнение — против продукта.

Стандарт (заморожен, общий): «полно» = каждое поведение брифа покрыто
исполняемым сценарием с точными наблюдаемыми; «качественно» = пять
деливераблов: спека, код, сценарии, двойной зелёный прогон, покрытие.

Три продукта (хост): <пути к каталогам рук задачи>. Продукты: P1 —
solo-product/ (сьют solo-product/check.sh); P2/P3 — продукт в корне,
сьют: экспортированная таблица кейсов руки
(.punchtape/canon/scenarios/*.yaml — записи: seed [{path,content}],
materials [{from}], pre [argv], when.command [argv], then
[наблюдения]; единый раннер campaigns/campaign-03/tools/case-runner.py).
Материал: <путь к материалам смысла> (read-only). Образ:
punchtape-world:th3.

Для КАЖДОГО продукта одна процедура:
1. Сьют×2 на собственном продукте во временных контейнерах
   (--network none):
   - P1: cd solo-product && sh check.sh — дважды в свежих контейнерах,
     оба зелёные, exit 0, выводы идентичны;
   - P2/P3: единый раннер по записям кейсов: свежий каталог на кейс,
     fixtures байт-в-байт, при необходимости собрать продукт
     (cargo build --offline / make / go build / node / PYTHONPATH),
     выполнить when.command, timeout 15, сверка then: exit/stdout/
     stderr/disk точным сравнением; расхождение только хвостовые
     \n/пробелы — pass с пометкой tolerant. Таблицу целиком дважды.
2. Деливераблы 0–5 одним методом: spec (SPEC.md / SPECIFICATION.md
   +specdocs); code (исходники, собирается/запускается); scenarios
   (SCENARIOS+check.sh / таблица кейсов); double_green (п.1);
   coverage (COVERAGE.md / specdocs-связка [SCN-xxx]). Недостающий —
   строка причины.
3. method-строка (как считал).

Результат: measure.json в каталог каждой руки:
{"suite_double_green": bool, "suite_method": "...",
 "suite_runs": [{"run":1,"pass":N,"fail":M,"tolerant":K},{...}],
 "deliverables": {"spec":1/0,"code":1/0,"scenarios":1/0,
 "double_green":1/0,"coverage":1/0}, "deliverables_count": 0-5,
 "coverage_tooling": "n/a|%," "notes": ["причины пробелов"]}

Закончи ход сводкой по трём продуктам: сьют×2 (счёт прогонов),
деливераблы n/5, причины пробелов.
