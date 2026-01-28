export interface MangaData {
  id: string
  title: string
  cover: string
  chapters: ChapterData[]
  total_pages?: number
}

export interface ChapterData {
  chapter_id: string
  chapter: number | string
  chapter_title?: string
  chapter_volume?: string
  group_name: string
  language: string
  time: string
}

export interface ChapterPages {
  pages: string[]
}
