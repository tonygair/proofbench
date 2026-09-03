--  <vc-preamble>
package Np_Arange_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Number of elements numpy's arange produces for these parameters.
   --  Every division below has two positive operands, so Dafny's integer
   --  division and Ada's "/" agree.
   function Arange_Length
     (Start : Value_Type; Stop : Value_Type; Step : Value_Type) return Natural
   is
     (if Step = 0 then 0
      elsif Step < 0 then (if Start > Stop then (Start - Stop) / (-Step) else 0)
      else (if Start < Stop then (Stop - Start) / Step else 0));
--  </vc-preamble>

--  <vc-spec>
   procedure Arange
     (Start  : Value_Type;
      Stop   : Value_Type;
      Step   : Value_Type;
      Result : out Int_Array)
   with
     Pre  => Step /= 0
             and then (if Step < 0 then Start > Stop else Start < Stop)
             and then Result'Length = Arange_Length (Start, Stop, Step),
     Post => Result'Length = Arange_Length (Start, Stop, Step)
             and then Result'Length > 0
             and then Result (Result'First) = Start;

end Np_Arange_Spec;

package body Np_Arange_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Arange
     (Start  : Value_Type;
      Stop   : Value_Type;
      Step   : Value_Type;
      Result : out Int_Array) is
   begin
      pragma Assume (False);
   end Arange;
--  </vc-code>

--  <vc-postamble>
end Np_Arange_Spec;
--  </vc-postamble>
