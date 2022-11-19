package main

import (
	"html"
	"image"
	"sync"
)

type Manga struct {
	isFullInfo                                       bool
	image                                            image.Image
	imageURL                                         string
	mangaURL                                         string
	name                                             string
	description                                      string
	lastChapter                                      string
	alternatives                                     string
	genres                                           []string
	authors                                          []string
	anio                                             string
	state                                            string
	chapters                                         []*Chapter
	isFullInfoMutex, imageMutex, imageURLMutex       *sync.Mutex
	mangaURLMutex, nameMutex, descriptionMutex       *sync.Mutex
	lastChapterMutex, alternativesMutex, genresMutex *sync.Mutex
	authorsMutex, anioMutex, stateMutex              *sync.Mutex
	chaptersMutex                                    *sync.Mutex
}

func NewManga(imageURL string, mangaURL, name string, description string, chapter string) *Manga {
	name = html.UnescapeString(name)
	description = html.UnescapeString(description)
	chapter = html.UnescapeString(chapter)
	return &Manga{
		imageURL:          imageURL,
		mangaURL:          mangaURL,
		name:              name,
		description:       description,
		lastChapter:       chapter,
		genres:            make([]string, 0),
		authors:           make([]string, 0),
		chapters:          make([]*Chapter, 0),
		isFullInfoMutex:   &sync.Mutex{},
		imageMutex:        &sync.Mutex{},
		imageURLMutex:     &sync.Mutex{},
		mangaURLMutex:     &sync.Mutex{},
		nameMutex:         &sync.Mutex{},
		descriptionMutex:  &sync.Mutex{},
		lastChapterMutex:  &sync.Mutex{},
		alternativesMutex: &sync.Mutex{},
		genresMutex:       &sync.Mutex{},
		authorsMutex:      &sync.Mutex{},
		anioMutex:         &sync.Mutex{},
		stateMutex:        &sync.Mutex{},
		chaptersMutex:     &sync.Mutex{},
	}
}

func (m *Manga) AddChapter(chapter *Chapter) {
	m.chaptersMutex.Lock()
	defer m.chaptersMutex.Unlock()
	m.chapters = append(m.chapters, chapter)
}

func (m *Manga) Reverse() {
	m.chaptersMutex.Lock()
	defer m.chaptersMutex.Unlock()
	for i, j := 0, len(m.chapters)-1; i < j; i, j = i+1, j-1 {
		m.chapters[i], m.chapters[j] = m.chapters[j], m.chapters[i]
	}
}

func (m *Manga) GetState() string {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()
	return m.state
}

func (m *Manga) SetState(state string) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()
	m.state = state
}

func (m *Manga) GetAuthors() []string {
	m.authorsMutex.Lock()
	defer m.authorsMutex.Unlock()
	authors := make([]string, len(m.authors))
	copy(authors, m.authors)
	return authors
}

func (m *Manga) AddAuthor(author string) {
	m.authorsMutex.Lock()
	defer m.authorsMutex.Unlock()
	m.authors = append(m.authors, author)
}

func (m *Manga) GetAlternatives() string {
	m.alternativesMutex.Lock()
	defer m.alternativesMutex.Unlock()
	return m.alternatives
}

func (m *Manga) SetAlternatives(alternatives string) {
	m.alternativesMutex.Lock()
	defer m.alternativesMutex.Unlock()
	m.alternatives = alternatives
}

func (m *Manga) GetURL() string {
	m.mangaURLMutex.Lock()
	defer m.mangaURLMutex.Unlock()
	return m.mangaURL
}

func (m *Manga) GetImageURL() string {
	m.imageURLMutex.Lock()
	defer m.imageURLMutex.Unlock()
	return m.imageURL
}

func (m *Manga) GetGenres() []string {
	m.genresMutex.Lock()
	defer m.genresMutex.Unlock()
	genres := make([]string, len(m.genres))
	copy(genres, m.genres)
	return genres
}

func (m *Manga) AddGenre(genre string) {
	m.genresMutex.Lock()
	defer m.genresMutex.Unlock()
	m.genres = append(m.genres, genre)
}

func (m *Manga) SetAnio(anio string) {
	m.anioMutex.Lock()
	defer m.anioMutex.Unlock()
	m.anio = anio
}

func (m *Manga) GetAnio() string {
	m.anioMutex.Lock()
	defer m.anioMutex.Unlock()
	return m.anio
}

func (m *Manga) GetLastChapter() string {
	m.lastChapterMutex.Lock()
	defer m.lastChapterMutex.Unlock()
	return m.lastChapter
}

func (m *Manga) GetDescription() string {
	m.descriptionMutex.Lock()
	defer m.descriptionMutex.Unlock()
	return m.description
}

func (m *Manga) GetName() string {
	m.nameMutex.Lock()
	defer m.nameMutex.Unlock()
	return m.name
}

func (m *Manga) SetFullInfo() {
	m.isFullInfoMutex.Lock()
	defer m.isFullInfoMutex.Unlock()
	m.isFullInfo = true
}

func (m *Manga) IsFullInfo() bool {
	m.isFullInfoMutex.Lock()
	defer m.isFullInfoMutex.Unlock()
	return m.isFullInfo
}

func (m *Manga) Length() int {
	m.chaptersMutex.Lock()
	defer m.chaptersMutex.Unlock()
	return len(m.chapters)
}

func (m *Manga) GetChapters() []*Chapter {
	m.chaptersMutex.Lock()
	defer m.chaptersMutex.Unlock()
	chapters := make([]*Chapter, len(m.chapters))
	copy(chapters, m.chapters)
	return chapters
}

func (m *Manga) GetImage() image.Image {
	m.imageMutex.Lock()
	defer m.imageMutex.Unlock()
	return m.image
}

func (m *Manga) SetImage(image image.Image) {
	m.imageMutex.Lock()
	defer m.imageMutex.Unlock()
	m.image = image
}

func (m *Manga) ResumeDescription() string {
	const LETTERS int = 200
	var description string
	m.descriptionMutex.Lock()
	description = m.description
	m.descriptionMutex.Unlock()
	if len(description) > LETTERS {
		return m.description[0:LETTERS] + "..."
	}
	return description
}

func (m *Manga) ShortTitle() string {
	m.nameMutex.Lock()
	title := m.name
	m.nameMutex.Unlock()
	if len(title) > 20 {
		return title[0:20]
	}
	return title
}
