## Gemini Crawler
A BFS crawler to index geminispace.

### Usage
`AWS_EFS_MOUNT` variable is used to set the location of the index. The default seed url is [gemini://gemini.circumlunar.space/](gemini://gemini.circumlunar.space/) but a `seeds.txt` can be provided with
one URL in each line.
```
go build
AWS_EFS_MOUNT=. ./gemini-crawler
```
The crawler will recursively find gemini links from each page and index the contents using [Bleve](https://github.com/blevesearch/bleve).