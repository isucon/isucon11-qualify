import axios, { AxiosRequestConfig } from 'axios'
import { dateToTimestamp, timestampToDate } from './date'

class Apis {
  async postAuth(jwt: string, axiosConfig?: AxiosRequestConfig) {
    await axios.post<void>(
      '/api/auth',
      {},
      {
        headers: { Authorization: `Bearer ${jwt}` },
        ...axiosConfig
      }
    )
  }

  async postSignout(axiosConfig?: AxiosRequestConfig) {
    await axios.post<void>('/api/signout', axiosConfig)
  }

  async getUserMe(axiosConfig?: AxiosRequestConfig) {
    const { data } = await axios.get<User>('/api/user/me', axiosConfig)
    return data
  }

  async getIsus(options?: { limit: number }, axiosConfig?: AxiosRequestConfig) {
    const { data } = await axios.get<Isu[]>(`/api/isu`, {
      params: options,
      ...axiosConfig
    })
    return data
  }

  async postIsu(req: PostIsuRequest, axiosConfig?: AxiosRequestConfig) {
    const data = new FormData()
    data.append('jia_isu_uuid', req.jia_isu_uuid)
    data.append('isu_name', req.isu_name)
    if (req.image) {
      data.append('image', req.image, req.image.name)
    }
    await axios.post<void>(`/api/isu`, data, {
      headers: { 'content-type': 'multipart/form-data' },
      ...axiosConfig
    })
  }

  async getIsuSearch(
    option?: IsuSearchRequest,
    axiosConfig?: AxiosRequestConfig
  ) {
    const { data } = await axios.get<Isu[]>(`/api/isu/search`, {
      params: option,
      ...axiosConfig
    })
    return data
  }

  async getIsu(jiaIsuUuid: string, axiosConfig?: AxiosRequestConfig) {
    const { data } = await axios.get<Isu>(`/api/isu/${jiaIsuUuid}`, axiosConfig)
    return data
  }

  async deleteIsu(jiaIsuUuid: string, axiosConfig?: AxiosRequestConfig) {
    await axios.delete<Isu>(`/api/isu/${jiaIsuUuid}`, axiosConfig)
  }

  async getIsuGraphs(
    jiaIsuUuid: string,
    req: GraphRequest,
    axiosConfig?: AxiosRequestConfig
  ) {
    const params: ApiGraphRequest = {
      datetime: dateToTimestamp(req.date)
    }
    const { data } = await axios.get<ApiGraph[]>(
      `/api/isu/${jiaIsuUuid}/graph`,
      {
        params,
        ...axiosConfig
      }
    )

    const graphs: Graph[] = []
    data.forEach(apiGraph => {
      graphs.push({
        ...apiGraph,
        start_at: timestampToDate(apiGraph.start_at),
        end_at: timestampToDate(apiGraph.end_at)
      })
    })
    return graphs
  }

  async getIsuConditions(
    jiaIsuUuid: string,
    req: ConditionRequest,
    axiosConfig?: AxiosRequestConfig
  ) {
    const params: ApiConditionRequest = {
      ...req,
      start_time: req.start_time ? dateToTimestamp(req.start_time) : undefined,
      end_time: dateToTimestamp(req.end_time)
    }
    const { data } = await axios.get<ApiCondition[]>(
      `/api/condition/${jiaIsuUuid}`,
      { params, ...axiosConfig }
    )

    const conditions: Condition[] = []
    data.forEach(apiCondition => {
      conditions.push({
        ...apiCondition,
        date: timestampToDate(apiCondition.timestamp)
      })
    })

    return conditions
  }
}

const apis = new Apis()
export default apis

export interface User {
  jia_user_id: string
}

export interface Isu {
  jia_isu_uuid: string
  name: string
  character: string
}

export interface IsuLog {
  jia_isu_uuid: string
  timestamp: number
  is_sitting: boolean
  condition: string
  message: string
  created_at: string
}

export interface GraphData {
  score: number
  sitting: number
  detail: { [key: string]: number }
}

interface ApiGraph {
  jia_isu_uuid: string
  start_at: number
  end_at: number
  data: GraphData | null
}

export interface Graph {
  jia_isu_uuid: string
  start_at: Date
  end_at: Date
  data: GraphData | null
}

export interface IsuSearchRequest {
  name?: string
  charactor?: string
  catalog_name?: string
  min_limit_weight?: number
  max_limit_weight?: number
  catalog_tags?: string
  page?: string
}

export const DEFAULT_SEARCH_LIMIT = 20

export interface PostIsuRequest {
  jia_isu_uuid: string
  isu_name: string
  image?: File
}

interface ApiCondition {
  jia_isu_uuid: string
  isu_name: string
  timestamp: number
  is_sitting: boolean
  condition: string
  condition_level: ConditionLevel
  message: string
}

export interface Condition {
  jia_isu_uuid: string
  isu_name: string
  date: Date
  is_sitting: boolean
  condition: string
  condition_level: ConditionLevel
  message: string
}

type ConditionLevel = 'info' | 'warning' | 'critical'

interface ApiConditionRequest {
  start_time?: number
  end_time: number
  // critical,warning,info をカンマ区切りで取り扱う
  condition_level: string
}

export interface ConditionRequest {
  start_time?: Date
  end_time: Date
  // critical,warning,info をカンマ区切りで取り扱う
  condition_level: string
}

interface ApiGraphRequest {
  datetime: number
}

export interface GraphRequest {
  date: Date
}

export const DEFAULT_CONDITION_LIMIT = 20
