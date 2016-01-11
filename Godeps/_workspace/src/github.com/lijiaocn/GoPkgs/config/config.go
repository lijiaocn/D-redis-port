//create: 2015/09/16 14:55:07 change: 2015/09/16 17:34:59 author:lijiao
package config
import(
	"encoding/json"
	"os"
)

func LoadConfig(jsonfile string, config interface{}) error{
	fi,err := os.Lstat(jsonfile)
	if err != nil{
		return err
	}
	buf := make([]byte, fi.Size(), fi.Size()+1)

	f,err := os.Open(jsonfile)
	defer f.Close()
	if err != nil{
		return err
	}

	n,err := f.Read(buf)
	if err != nil{
		return err
	}

	err = json.Unmarshal(buf[0:n], config)
	if err != nil{
		return err
	}
	return nil
}

func GenConfig(config interface{}, outfile string) ([]byte, error){
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil{
		return nil, err
	}

	if len(outfile)!= 0{
		f,err := os.OpenFile(outfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil{
			return b, err
		}
		_, err = f.Write(b)
		if err != nil{
			return b, err
		}
		err = f.Close()
		if err != nil{
			return b, err
		}
	}
	return b, err
}

