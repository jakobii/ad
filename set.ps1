function New-Credential {
    param(
        [parameter(Mandatory = $true)]
        [alias("u")]
        [string]
        $Username,

        [parameter(Mandatory = $true)]
        [alias("p")]
        [string]
        $Password
    )
    $SecureString = ConvertTo-SecureString $Password -AsPlainText -Force
    return [System.Management.Automation.PSCredential]::new($Username, $SecureString)
}

function Test-Credential {
    param(
        [string]$Server,
        [string]$Domain,
        [string]$Username,
        [string]$Password
    )

    # build the ldap uri
    [string[]]$DomainParts = $Domain -split '[.]'
    [string[]]$Path = @()
    foreach($part in $DomainParts){
        $Path += "DC=" + $part
    }
    $uri = "LDAP://$Server/" + $($Path -join ",")

    # attempt connection
    $result = [DirectoryServices.DirectoryEntry]::new($uri ,$Username ,$Password)
    
    if([convert]::ToString($result.Name) -eq ''){
        return $false
    }
    return $true
}

#Test-Credential -Server '192.168.56.99' -Domain 'lab.local' -Username 'tim' -Password 'abcd1234!'

Get-ADDomain -Server '192.168.56.99' -Credential $(New-Credential -Username 'tim' -Password 'abcd1234!')

Get-ADUser -Server '192.168.56.99' -Credential $(New-Credential -Username 'tim' -Password 'abcd1234!') -Identity 'tim'

$(
function New-Credential {
    param(
        [parameter(Mandatory)][string]$Username,
        [parameter(Mandatory)][string]$Password
    )
    $SecureString = ConvertTo-SecureString $Password -AsPlainText -Force
    return [System.Management.Automation.PSCredential]::new($Username, $SecureString)
};
New-Credential -Username 'tim' -Password 'abcd1234!'
)

