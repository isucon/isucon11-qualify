import { useState } from 'react'
import { useHistory } from 'react-router-dom'
import Card from '/@/components/UI/Card'
import UploadImageButton from '/@/components/Main/UploadImageButton'
import Input from '/@/components/UI/Input'
import apis, { PostIsuRequest } from '/@/lib/apis'
import Button from '/@/components/UI/Button'
import toast from 'react-hot-toast'

const Register = () => {
  const [id, setId] = useState('')
  const [name, setName] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const history = useHistory()

  const submit = async () => {
    try {
      const req: PostIsuRequest = {
        jia_isu_uuid: id,
        isu_name: name
      }
      if (file) {
        req.image = file
      }
      await apis.postIsu(req)
      history.push(`/isu/${id}`)
    } catch (e) {
      if (e.response.status === 409) {
        history.push(`/isu/${id}`)
      } else {
        toast.error(e.response.data)
      }
    }
  }

  return (
    <div className="flex justify-center p-10">
      <div className="flex justify-center w-full max-w-2xl">
        <Card>
          <div className="w-full">
            <h2 className="mb-8 text-xl font-bold">ISUを登録</h2>
            <div className="flex flex-col gap-4">
              <Input label="JIAのIsuID" value={id} setValue={setId} />
              <Input label="ISUの名前" value={name} setValue={setName} />
              <div className="flex flex-col gap-8 items-center mt-6">
                <UploadImageButton putIsuIcon={setFile} />
                <Button
                  label="登録"
                  onClick={submit}
                  customClass="px-4 py-1 h-8 text-white font-bold bg-button rounded-2xl"
                  disabled={!id || !name}
                />
              </div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}

export default Register
