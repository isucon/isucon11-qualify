import { FormEvent, useState } from 'react'
import Input from '../components/UI/Input'
import apis from '../lib/apis'

const Register = () => {
  const [id, setId] = useState('')
  const [name, setName] = useState('')
  const submit = (event: FormEvent<HTMLFormElement>) => {
    apis.postIsu({ jia_isu_uuid: id, isu_name: name })
    event.preventDefault()
  }

  return (
    <div>
      <form onSubmit={submit}>
        <Input label={'JIAのIsuID'} value={id} setValue={setId}></Input>
        <Input label={'ISUの名前'} value={name} setValue={setName}></Input>
        <button type="submit" className="border-dark-200 border">
          登録
        </button>
      </form>
    </div>
  )
}

export default Register
