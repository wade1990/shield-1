package agent

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AuthorizedKeysFile string `yaml:"authorized_keys_file"`
	HostKeyFile string `yaml:"host_key_file"`
	ListenAddress string `yaml:"listen_address"`
	PluginPaths []string `yaml:"plugin_paths"`
}

func (agent *Agent) ReadConfig(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var config Config
	err = yaml.Unmarshal(b, &config)

	if err != nil {
		return err
	}

	authorizedKeys, err := LoadAuthorizedKeys(config.AuthorizedKeysFile)
	if err != nil {
		fmt.Printf("failed to load authorized keys: %s\n", err)
		return err
	}

	server, err := ConfigureSSHServer(config.HostKeyFile, authorizedKeys)
	if err != nil {
		fmt.Printf("failed to configure SSH server: %s\n", err)
		return err
	}
	agent.config = server

	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		fmt.Printf("failed to bind %s: %s", config.ListenAddress, err)
		return err
	}
	agent.Listen = listener

	fmt.Printf("listening on %s\n", config.ListenAddress)

	agent.PluginPaths = config.PluginPaths

	return nil
}

func LoadAuthorizedKeys(path string) ([]ssh.PublicKey, error) {
	authorizedKeysBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var authorizedKeys []ssh.PublicKey

	for {
		key, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			break
		}

		authorizedKeys = append(authorizedKeys, key)

		authorizedKeysBytes = rest
	}

	return authorizedKeys, nil
}

func ConfigureSSHServer(hostKeyPath string, authorizedKeys []ssh.PublicKey) (*ssh.ServerConfig, error) {
	certChecker := &ssh.CertChecker{
		IsAuthority: func(key ssh.PublicKey) bool {
			return false
		},

		UserKeyFallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			for _, k := range authorizedKeys {
				if bytes.Equal(k.Marshal(), key.Marshal()) {
					return nil, nil
				}
			}

			return nil, fmt.Errorf("unknown public key")
		},
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return certChecker.Authenticate(conn, key)
		},
	}

	privateBytes, err := ioutil.ReadFile(hostKeyPath)
	if err != nil {
		return nil, err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, err
	}

	config.AddHostKey(private)

	return config, nil
}

func ConfigureSSHClient(privateKeyPath string) (*ssh.ClientConfig, error) {
	raw, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(raw)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}, nil
}
