import { DatabaseService } from '../../bindings/mangav5/services'

export const getMangaDirectory = async () => {
  try {
    const config = await DatabaseService.GetConfig('manga_directory')
    return config?.Value
  } catch (error) {
    console.log(error)
  }
}

export const setMangaDirectory = async (directory: string) => {
  try {
    await DatabaseService.SetConfig('manga_directory', directory)
  } catch (error) {
    console.log(error)
  }
}
