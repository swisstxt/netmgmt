%global service_name %{name}

Name:             %{name}
Version:          %{ver}
Release:          %{rel}%{?dist}
Summary:          netmgmt for RHEL/CENTOS %{os_rel}
BuildArch:        %{arch}
Group:            Application/Internet
License:          commercial
URL:              https://github.com/swisstxt/netmgmt
BuildRoot:        %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

Source1:        netmgmt.bin
Source2:        netdef.yaml
Source3:        netmgmt.service
Source4:        netmgmt.sysconfig

%define appdir /opt/%{name}
%define systemd_dest /usr/lib/systemd/system/
%define sysconfig /etc/sysconfig/

%description
netmgmt for RHEL/CENTOS %{os_rel}

%prep

%build

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/%{appdir}
mkdir -p $RPM_BUILD_ROOT/%{systemd_dest}
mkdir -p $RPM_BUILD_ROOT/%{sysconfig}
%{__install} -p -m 0755 %{SOURCE1} $RPM_BUILD_ROOT/%{appdir}/netmgmt
%{__install} -p -m 0750 %{SOURCE2} $RPM_BUILD_ROOT/%{appdir}/netdef.yaml
%{__install} -p -m 0755 %{SOURCE3} $RPM_BUILD_ROOT/%{systemd_dest}/netmgmt.service
%{__install} -p -m 0755 %{SOURCE4} $RPM_BUILD_ROOT/%{sysconfig}/netmgmt

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root)
%attr(0755,root,root) %{appdir}
%attr(0755,root,root) %{appdir}/*
%attr(0755,root,root) %{systemd_dest}/netmgmt.service
%config(noreplace) %{appdir}/netdef.yaml
%config(noreplace) %{sysconfig}/netmgmt

%changelog
* Thu Jan 12 2015 Daniel Menet <daniel.menet@swisstxt.ch> - 1-1
Initial creation
